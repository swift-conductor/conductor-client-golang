//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package worker_e2e

import (
	"github.com/conductor-sdk/conductor-go/sdk/model"
	"os"
	"testing"
	"time"

	"github.com/conductor-sdk/conductor-go/tests/e2e/e2e_properties"
	log "github.com/sirupsen/logrus"
)

const (
	taskName = "TEST_GO_TASK_SIMPLE"

	workflowName              = "TEST_GO_WORKFLOW_SIMPLE"
	workflowCompletionTimeout = 5 * time.Second
	workflowExecutionQty      = 15

	workerQty          = 7
	workerPollInterval = 250 * time.Millisecond
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func TestWorkers(t *testing.T) {
	outputData := map[string]interface{}{
		"key": "value",
	}
	workerWithTaskResultOutput := func(t *model.Task) (interface{}, error) {
		taskResult := model.NewTaskResultFromTask(t)
		taskResult.OutputData = outputData
		taskResult.Status = model.CompletedTask
		return taskResult, nil
	}
	workerWithGenericOutput := func(t *model.Task) (interface{}, error) {
		return outputData, nil
	}
	workers := []model.ExecuteTaskFunction{
		workerWithTaskResultOutput,
		workerWithGenericOutput,
	}
	for _, worker := range workers {
		err := validateWorker(worker, outputData)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func validateWorker(worker model.ExecuteTaskFunction, expectedOutput map[string]interface{}) error {
	workflowIdList, err := e2e_properties.StartWorkflows(
		workflowExecutionQty,
		workflowName,
	)
	if err != nil {
		return err
	}
	err = e2e_properties.TaskRunner.StartWorker(
		taskName,
		worker,
		workerQty,
		workerPollInterval,
	)
	if err != nil {
		return err
	}
	runningWorkflows := make([]chan error, len(workflowIdList))
	for i, workflowId := range workflowIdList {
		runningWorkflows[i] = make(chan error)
		go e2e_properties.ValidateWorkflowDaemon(
			workflowCompletionTimeout,
			runningWorkflows[i],
			workflowId,
			expectedOutput,
		)
	}
	for _, runningWorkflowChannel := range runningWorkflows {
		err := <-runningWorkflowChannel
		if err != nil {
			return err
		}
	}
	return e2e_properties.TaskRunner.RemoveWorker(
		taskName,
		workerQty,
	)
}
