//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package worker

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/swift-conductor/conductor-client-golang/sdk/client"
	"github.com/swift-conductor/conductor-client-golang/sdk/concurrency"
	"github.com/swift-conductor/conductor-client-golang/sdk/metrics"
	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/settings"

	"github.com/antihax/optional"
	log "github.com/sirupsen/logrus"
)

const taskUpdateRetryAttemptsLimit = 3

var (
	sleepForOnNoAvailableWorker = 10 * time.Millisecond
	sleepForOnGenericError      = 200 * time.Millisecond
)

var hostname, _ = os.Hostname()

// WorkerHost implements polling and execution logic for a Conductor worker. Every polling interval, each running
// task attempts to retrieve a from Conductor. Multiple tasks can be started in parallel. All Goroutines started by this
// worker cannot be stopped, only paused and resumed.
//
// Conductor tasks are tracked by name separately. Each WorkerHost tracks a separate poll interval and batch size for
// each task, which is shared by all workers running that task. For instance, if task "foo" is running with a batch size
// of n, and k workers, the average number of tasks retrieved during each polling interval is n*k.
//
// All methods on WorkerHost are thread-safe.
type WorkerHost struct {
	conductorTaskResourceClient *client.TaskResourceApiService

	workerWaitGroup sync.WaitGroup

	batchSizeByTaskNameMutex sync.RWMutex
	batchSizeByTaskName      map[string]int

	runningWorkersByTaskNameMutex sync.RWMutex
	runningWorkersByTaskName      map[string]int

	pollIntervalByTaskNameMutex sync.RWMutex
	pollIntervalByTaskName      map[string]time.Duration

	pausedWorkersMutex sync.RWMutex
	pausedWorkers      map[string]bool
}

// NewWorkerHost returns a new WorkerHost using the provided settings.
func NewWorkerHost(httpSettings *settings.HttpSettings) *WorkerHost {
	apiClient := client.NewAPIClient(
		httpSettings,
	)
	return NewWorkerHostWithApiClient(apiClient)
}

// NewWorkerHostWithApiClient creates a new WorkerHost which uses the provided client.APIClient to communicate with
// Conductor.
func NewWorkerHostWithApiClient(apiClient *client.APIClient) *WorkerHost {
	return &WorkerHost{
		conductorTaskResourceClient: &client.TaskResourceApiService{
			APIClient: apiClient,
		},
		batchSizeByTaskName:      make(map[string]int),
		runningWorkersByTaskName: make(map[string]int),
		pollIntervalByTaskName:   make(map[string]time.Duration),
		pausedWorkers:            make(map[string]bool),
	}
}

// SetSleepOnGenericError Sets the time for which to wait before continuing to poll/execute when there is an error
// Default is 200 millis, and this function can be used to increase/decrease the duration of the wait time
// Useful to avoid excessive logs in the worker when there are intermittent issues
func (c *WorkerHost) SetSleepOnGenericError(duration time.Duration) {
	sleepForOnGenericError = duration
}

// StartWorkerWithDomain starts a polling worker on a new goroutine, which only polls for tasks using the provided
// domain. Equivalent to:
//
//	StartWorkerWithDomain(taskName, executeFunction, batchSize, pollInterval, "")
func (c *WorkerHost) StartWorkerWithDomain(taskName string, executeFunction model.WorkerTaskFunction, batchSize int, pollInterval time.Duration, domain string) error {
	return c.startWorker(taskName, executeFunction, batchSize, pollInterval, domain)
}

// StartWorker starts a worker on a new goroutine, which polls conductor periodically for tasks matching the provided
// taskName and, if any are available, uses executeFunction to run them on a separate goroutine. Each call to
// StartWorker starts a new goroutine which performs batch polling to retrieve as many
// tasks from Conductor as are available, up to the batchSize set for the task. This func additionally sets the
// pollInterval and increases the batch size for the task, which applies to all tasks shared by this WorkerHost with the
// same taskName.
func (c *WorkerHost) StartWorker(taskName string, executeFunction model.WorkerTaskFunction, batchSize int, pollInterval time.Duration) error {
	return c.startWorker(taskName, executeFunction, batchSize, pollInterval, "")
}

// SetBatchSize can be used to set the batch size for all workers running the provided task.
func (c *WorkerHost) SetBatchSize(taskName string, batchSize int) error {
	if batchSize < 0 {
		return fmt.Errorf("batchSize can not be negative")
	}
	if !c.isWorkerRegistered(taskName) {
		return fmt.Errorf("no worker registered for taskName: %s", taskName)
	}
	c.batchSizeByTaskNameMutex.Lock()
	defer c.batchSizeByTaskNameMutex.Unlock()
	previous := c.batchSizeByTaskName[taskName]
	c.batchSizeByTaskName[taskName] = batchSize
	log.Debug(
		"Set batchSize for task: ", taskName,
		", from: ", previous,
		", to: ", c.batchSizeByTaskName[taskName],
	)
	if batchSize == 0 {
		log.Info("Stopped worker for task: ", taskName)
	} else if previous == 0 && c.batchSizeByTaskName[taskName] > 0 {
		log.Info("Started worker for task: ", taskName)
	}
	return nil
}

// IncreaseBatchSize increases the batch size used for all workers running the provided task.
func (c *WorkerHost) IncreaseBatchSize(taskName string, batchSize int) error {
	if batchSize < 1 {
		return fmt.Errorf("batchSize value must be positive")
	}
	if !c.isWorkerRegistered(taskName) {
		return fmt.Errorf("no worker registered for taskName: %s", taskName)
	}
	c.batchSizeByTaskNameMutex.Lock()
	defer c.batchSizeByTaskNameMutex.Unlock()
	previous := c.batchSizeByTaskName[taskName]
	c.batchSizeByTaskName[taskName] += batchSize
	log.Debug(
		"Increased batchSize for task: ", taskName,
		", from: ", previous,
		", to: ", c.batchSizeByTaskName[taskName],
	)
	if previous == 0 {
		log.Info("Started worker for task: ", taskName)
	}
	return nil
}

// DecreaseBatchSize decreases the batch size used for all workers running the provided task.
func (c *WorkerHost) DecreaseBatchSize(taskName string, batchSize int) error {
	if batchSize < 1 {
		return fmt.Errorf("batchSize value must be positive")
	}
	if !c.isWorkerRegistered(taskName) {
		return fmt.Errorf("no worker registered for taskName: %s", taskName)
	}
	c.batchSizeByTaskNameMutex.Lock()
	defer c.batchSizeByTaskNameMutex.Unlock()
	previous := c.batchSizeByTaskName[taskName]
	c.batchSizeByTaskName[taskName] -= batchSize
	log.Debug(
		"Decreased batchSize for task: ", taskName,
		", from: ", previous,
		", to: ", c.batchSizeByTaskName[taskName],
	)
	if previous-batchSize <= 0 {
		c.batchSizeByTaskName[taskName] = 0
		log.Info("Stopped worker for task: ", taskName)
	}
	return nil
}

// Pause pauses all workers running the provided task. When paused, workers will not poll for new tasks and no new
// goroutines are started. However it does not stop any goroutines running. Workers must be resumed at a later time
// using Resume. Failing to call `Resume()` on a WorkerHost running one or more workers can result in a goroutine leak.
func (c *WorkerHost) Pause(taskName string) {
	c.pausedWorkersMutex.Lock()
	defer c.pausedWorkersMutex.Unlock()
	c.pausedWorkers[taskName] = true
}

// Resume all running workers for the provided taskName. If workers for the provided task are not paused, calling this
// method has no impact.
func (c *WorkerHost) Resume(taskName string) {
	c.pausedWorkersMutex.Lock()
	defer c.pausedWorkersMutex.Unlock()
	c.pausedWorkers[taskName] = false
}

func (c *WorkerHost) isPaused(taskName string) bool {
	c.pausedWorkersMutex.RLock()
	defer c.pausedWorkersMutex.RUnlock()
	return c.pausedWorkers[taskName]
}

// WaitWorkers uses an internal waitgroup to block the calling thread until all workers started by this WorkerHost have
// been stopped.
func (c *WorkerHost) WaitWorkers() {
	c.workerWaitGroup.Wait()
}

func (c *WorkerHost) startWorker(taskName string, executeFunction model.WorkerTaskFunction, batchSize int, pollInterval time.Duration, taskDomain string) error {
	c.SetPollIntervalForTask(taskName, pollInterval)
	c.Resume(taskName)
	previousMaxAllowedWorkers, err := c.getMaxAllowedWorkers(taskName)
	if err != nil {
		return err
	}
	err = c.increaseMaxAllowedWorkers(taskName, batchSize)
	if err != nil {
		return err
	}
	if previousMaxAllowedWorkers < 1 {
		c.workerWaitGroup.Add(1)
		go c.work4ever(taskName, executeFunction, taskDomain)
	}
	log.Info(
		fmt.Sprintf(
			"Started %d worker(s) for taskName %s, polling in interval of %d ms",
			batchSize,
			taskName,
			pollInterval.Milliseconds(),
		),
	)
	return nil
}

func (c *WorkerHost) work4ever(taskName string, executeFunction model.WorkerTaskFunction, domain string) {
	defer c.workerWaitGroup.Done()
	defer concurrency.HandlePanicError("poll_and_execute")
	for c.isWorkerRegistered(taskName) {
		c.workOnce(taskName, executeFunction, domain)
	}
}

func (c *WorkerHost) workOnce(taskName string, executeFunction model.WorkerTaskFunction, domain string) {
	if c.isPaused(taskName) {
		pauseOnGenericError(taskName, domain, fmt.Errorf("worker is paused"))
		return
	}
	batchSize, err := c.getAvailableWorkerAmount(taskName)
	if err != nil {
		pauseOnGenericError(
			taskName, domain,
			fmt.Errorf("failed to get the number of available workers, reason: %s", err.Error()),
		)
		return
	}
	if batchSize < 1 {
		pauseOnNoAvailableWorkerError(taskName, domain)
		return
	}
	tasks, err := c.batchPoll(taskName, batchSize, domain)
	if err != nil {
		pauseOnGenericError(
			taskName, domain,
			fmt.Errorf("failed to poll, reason: %s", err.Error()),
		)
		return
	}
	if len(tasks) < 1 {
		pollInterval, err := c.GetPollIntervalForTask(taskName)
		if err != nil {
			log.Error(err)
			pauseOnGenericError(
				taskName, domain,
				fmt.Errorf("failed to get poll interval, reason: %s", err.Error()),
			)
			return
		}
		time.Sleep(pollInterval)
		return
	}
	for _, task := range tasks {
		c.increaseRunningWorkers(taskName)
		go c.executeAndUpdateTask(taskName, task, executeFunction)
	}
}

func (c *WorkerHost) executeAndUpdateTask(taskName string, task model.WorkerTask, executeFunction model.WorkerTaskFunction) {
	defer c.runningWorkerDone(taskName)
	defer concurrency.HandlePanicError("execute_and_update_task " + string(task.TaskId) + ": " + string(task.Status))
	taskResult := c.executeTask(&task, executeFunction)
	err := c.updateTaskWithRetry(taskName, taskResult)
	if err != nil {
		log.Error("failed to update task ", taskName, ",taskId = ", task.TaskId, ",workflowId = ", task.WorkflowInstanceId, ",", err)
	}
}

func (c *WorkerHost) batchPoll(taskName string, count int, domain string) ([]model.WorkerTask, error) {
	timeout, err := c.GetPollIntervalForTask(taskName)
	if err != nil {
		return nil, err
	}
	var domainOptional optional.String
	if domain != "" {
		domainOptional = optional.NewString(domain)
	}
	log.Debug(
		"Polling for task: ", taskName,
		", in batches of size: ", count,
	)
	metrics.IncrementTaskPoll(taskName)
	startTime := time.Now()
	tasks, response, err := c.conductorTaskResourceClient.BatchPoll(
		context.Background(),
		taskName,
		&client.TaskResourceApiBatchPollOpts{
			Domain:   domainOptional,
			Workerid: optional.NewString(hostname),
			Count:    optional.NewInt32(int32(count)),
			Timeout:  optional.NewInt32(int32(timeout.Milliseconds())),
		},
	)
	spentTime := time.Since(startTime)
	metrics.RecordTaskPollTime(
		taskName,
		spentTime.Seconds(),
	)
	if err != nil {
		metrics.IncrementTaskPollError(
			taskName, err,
		)
		return nil, err
	}
	if response.StatusCode == 204 {
		return nil, nil
	}
	log.Debug(fmt.Sprintf("Polled %d tasks for taskName: %s", len(tasks), taskName))
	return tasks, nil
}

func (c *WorkerHost) executeTask(t *model.WorkerTask, executeFunction model.WorkerTaskFunction) *model.TaskResult {
	log.Trace(
		"Executing task of type: ", t.TaskDefName,
		", taskId: ", t.TaskId,
		", workflowId: ", t.WorkflowInstanceId,
	)
	startTime := time.Now()
	taskExecutionOutput, err := executeFunction(t)
	spentTime := time.Since(startTime)
	metrics.RecordTaskExecuteTime(
		t.TaskDefName, float64(spentTime.Milliseconds()),
	)
	if err != nil {
		metrics.IncrementTaskExecuteError(t.TaskDefName, err)
		log.Debug(
			"failed to execute task",
			", reason: ", err.Error(),
			", taskName: ", t.TaskDefName,
			", taskId: ", t.TaskId,
			", workflowId: ", t.WorkflowInstanceId,
		)
		if taskExecutionOutput == nil {
			return model.NewTaskResultFromTaskWithError(t, err)
		}
	}
	taskResult, err := model.GetTaskResultFromTaskExecutionOutput(t, taskExecutionOutput)
	if err != nil {
		log.Debug(
			"Failed to extract taskResult from generated object",
			", reason: ", err.Error(),
			", task type: ", t.TaskDefName,
			", taskId: ", t.TaskId,
			", workflowId: ", t.WorkflowInstanceId,
			", response: ", err,
		)
		return model.NewTaskResultFromTaskWithError(t, err)
	}
	log.Trace(
		"Executed task of type: ", t.TaskDefName,
		", taskId: ", t.TaskId,
		", workflowId: ", t.WorkflowInstanceId,
	)
	return taskResult
}

func (c *WorkerHost) updateTaskWithRetry(taskName string, taskResult *model.TaskResult) error {
	log.Debug(
		"Updating task of type: ", taskName,
		", taskId: ", taskResult.TaskId,
		", workflowId: ", taskResult.WorkflowInstanceId,
	)
	var lastError error
	for attempt := 0; attempt <= taskUpdateRetryAttemptsLimit; attempt += 1 {
		if attempt > 0 {
			// Wait for [10s, 20s, 30s] before next attempt
			amount := attempt * 10
			time.Sleep(time.Duration(amount) * time.Second)
		}
		_, err := c.updateTask(taskName, taskResult)
		if err == nil {
			log.Debug(
				"Updated task of type: ", taskName,
				", taskId: ", taskResult.TaskId,
				", workflowId: ", taskResult.WorkflowInstanceId,
			)
			return nil
		}
		metrics.IncrementTaskUpdateError(taskName, err)
		lastError = err
	}
	return fmt.Errorf("failed to update task %s after %d attempts. %s", taskName, taskUpdateRetryAttemptsLimit, lastError)
}

func (c *WorkerHost) updateTask(taskName string, taskResult *model.TaskResult) (*http.Response, error) {
	startTime := time.Now()
	_, response, err := c.conductorTaskResourceClient.UpdateTask(context.Background(), taskResult)
	spentTime := time.Since(startTime).Milliseconds()
	metrics.RecordTaskUpdateTime(taskName, float64(spentTime))
	return response, err
}

func (c *WorkerHost) getAvailableWorkerAmount(taskName string) (int, error) {
	allowed, err := c.getMaxAllowedWorkers(taskName)
	if err != nil {
		return -1, err
	}
	running, err := c.getRunningWorkers(taskName)
	if err != nil {
		return -1, err
	}
	return allowed - running, nil
}

func (c *WorkerHost) getMaxAllowedWorkers(taskName string) (int, error) {
	c.batchSizeByTaskNameMutex.RLock()
	defer c.batchSizeByTaskNameMutex.RUnlock()
	amount, ok := c.batchSizeByTaskName[taskName]
	if !ok {
		return 0, nil
	}
	return amount, nil
}

func (c *WorkerHost) getRunningWorkers(taskName string) (int, error) {
	c.runningWorkersByTaskNameMutex.RLock()
	defer c.runningWorkersByTaskNameMutex.RUnlock()
	amount, ok := c.runningWorkersByTaskName[taskName]
	if !ok {
		return 0, nil
	}
	return amount, nil
}

func (c *WorkerHost) isWorkerRegistered(taskName string) bool {
	c.batchSizeByTaskNameMutex.RLock()
	defer c.batchSizeByTaskNameMutex.RUnlock()
	_, ok := c.batchSizeByTaskName[taskName]
	return ok
}

func (c *WorkerHost) increaseRunningWorkers(taskName string) error {
	c.runningWorkersByTaskNameMutex.Lock()
	defer c.runningWorkersByTaskNameMutex.Unlock()
	c.runningWorkersByTaskName[taskName] += 1
	log.Trace("Increased running workers for task: ", taskName)
	return nil
}

func (c *WorkerHost) runningWorkerDone(taskName string) error {
	c.runningWorkersByTaskNameMutex.Lock()
	defer c.runningWorkersByTaskNameMutex.Unlock()
	c.runningWorkersByTaskName[taskName] -= 1
	log.Trace("Running worker done for task: ", taskName)
	return nil
}

func (c *WorkerHost) increaseMaxAllowedWorkers(taskName string, batchSize int) error {
	c.batchSizeByTaskNameMutex.Lock()
	defer c.batchSizeByTaskNameMutex.Unlock()
	c.batchSizeByTaskName[taskName] += batchSize
	log.Debug("Increased max allowed workers of task: ", taskName, ", by: ", batchSize)
	return nil
}

// SetPollIntervalForTask sets the pollInterval for all workers running the task with the provided taskName.
func (c *WorkerHost) SetPollIntervalForTask(taskName string, pollInterval time.Duration) error {
	c.pollIntervalByTaskNameMutex.Lock()
	defer c.pollIntervalByTaskNameMutex.Unlock()
	c.pollIntervalByTaskName[taskName] = pollInterval
	log.Info("Updated poll interval for task: ", taskName, ", to: ", pollInterval.Milliseconds(), "ms")
	return nil
}

// GetPollIntervalForTask retrieves the poll interval for all tasks running the provided taskName. An error is returned
// if no pollInterval has been registered for the provided task.
func (c *WorkerHost) GetPollIntervalForTask(taskName string) (pollInterval time.Duration, err error) {
	c.pollIntervalByTaskNameMutex.RLock()
	defer c.pollIntervalByTaskNameMutex.RUnlock()
	pollInterval, ok := c.pollIntervalByTaskName[taskName]
	if !ok {
		return pollInterval, fmt.Errorf("poll interval not registered for task: %s", taskName)
	}
	return pollInterval, nil
}

// GetBatchSizeForAll returns a map from taskName to batch size for all batch sizes currently registered with this
// WorkerHost.
func (c *WorkerHost) GetBatchSizeForAll() (batchSizeByTaskName map[string]int) {
	c.batchSizeByTaskNameMutex.RLock()
	defer c.batchSizeByTaskNameMutex.RUnlock()
	batchSizeByTaskName = make(map[string]int)
	for taskName, batchSize := range c.batchSizeByTaskName {
		batchSizeByTaskName[taskName] = batchSize
	}
	return batchSizeByTaskName
}

// GetBatchSizeForTask retrieves the current batch size for the provided task.
func (c *WorkerHost) GetBatchSizeForTask(taskName string) (batchSize int) {
	c.batchSizeByTaskNameMutex.RLock()
	defer c.batchSizeByTaskNameMutex.RUnlock()
	batchSize, ok := c.batchSizeByTaskName[taskName]
	if !ok {
		return 0
	}
	return batchSize
}

func pauseOnGenericError(taskName string, domain string, err error) {
	log.Error(fmt.Errorf("[%s][%s] %s", taskName, domain, err))
	time.Sleep(sleepForOnGenericError)
}

func pauseOnNoAvailableWorkerError(taskName string, domain string) {
	log.Trace(fmt.Errorf("no worker available for the task %s, domain %s", taskName, domain))
	time.Sleep(sleepForOnNoAvailableWorker)
}
