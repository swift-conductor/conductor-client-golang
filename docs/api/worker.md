<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# worker

```go
import "github.com/swift-conductor/conductor-client-golang/sdk/worker"
```

## Index

- [type WorkerRunner](<#WorkerRunner>)
  - [func NewWorkerRunner\(httpSettings \*settings.HttpSettings\) \*WorkerRunner](<#NewWorkerRunner>)
  - [func NewWorkerRunnerWithApiClient\(apiClient \*client.APIClient\) \*WorkerRunner](<#NewWorkerRunnerWithApiClient>)
  - [func \(c \*WorkerRunner\) DecreaseBatchSize\(taskName string, batchSize int\) error](<#WorkerRunner.DecreaseBatchSize>)
  - [func \(c \*WorkerRunner\) GetBatchSizeForAll\(\) \(batchSizeByTaskName map\[string\]int\)](<#WorkerRunner.GetBatchSizeForAll>)
  - [func \(c \*WorkerRunner\) GetBatchSizeForTask\(taskName string\) \(batchSize int\)](<#WorkerRunner.GetBatchSizeForTask>)
  - [func \(c \*WorkerRunner\) GetPollIntervalForTask\(taskName string\) \(pollInterval time.Duration, err error\)](<#WorkerRunner.GetPollIntervalForTask>)
  - [func \(c \*WorkerRunner\) IncreaseBatchSize\(taskName string, batchSize int\) error](<#WorkerRunner.IncreaseBatchSize>)
  - [func \(c \*WorkerRunner\) Pause\(taskName string\)](<#WorkerRunner.Pause>)
  - [func \(c \*WorkerRunner\) Resume\(taskName string\)](<#WorkerRunner.Resume>)
  - [func \(c \*WorkerRunner\) SetBatchSize\(taskName string, batchSize int\) error](<#WorkerRunner.SetBatchSize>)
  - [func \(c \*WorkerRunner\) SetPollIntervalForTask\(taskName string, pollInterval time.Duration\) error](<#WorkerRunner.SetPollIntervalForTask>)
  - [func \(c \*WorkerRunner\) SetSleepOnGenericError\(duration time.Duration\)](<#WorkerRunner.SetSleepOnGenericError>)
  - [func \(c \*WorkerRunner\) StartWorker\(taskName string, executeFunction model.WorkerTaskFunction, batchSize int, pollInterval time.Duration\) error](<#WorkerRunner.StartWorker>)
  - [func \(c \*WorkerRunner\) StartWorkerWithDomain\(taskName string, executeFunction model.WorkerTaskFunction, batchSize int, pollInterval time.Duration, domain string\) error](<#WorkerRunner.StartWorkerWithDomain>)
  - [func \(c \*WorkerRunner\) WaitWorkers\(\)](<#WorkerRunner.WaitWorkers>)


<a name="WorkerRunner"></a>
## type [WorkerRunner](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L48-L64>)

WorkerRunner implements polling and execution logic for a Conductor worker. Every polling interval, each running task attempts to retrieve a from Conductor. Multiple tasks can be started in parallel. All Goroutines started by this worker cannot be stopped, only paused and resumed.

Conductor tasks are tracked by name separately. Each WorkerRunner tracks a separate poll interval and batch size for each task, which is shared by all workers running that task. For instance, if task "foo" is running with a batch size of n, and k workers, the average number of tasks retrieved during each polling interval is n\*k.

All methods on WorkerRunner are thread\-safe.

```go
type WorkerRunner struct {
    // contains filtered or unexported fields
}
```

<a name="NewWorkerRunner"></a>
### func [NewWorkerRunner](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L67>)

```go
func NewWorkerRunner(httpSettings *settings.HttpSettings) *WorkerRunner
```

NewWorkerRunner returns a new WorkerRunner using the provided settings.

<a name="NewWorkerRunnerWithApiClient"></a>
### func [NewWorkerRunnerWithApiClient](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L76-L78>)

```go
func NewWorkerRunnerWithApiClient(apiClient *client.APIClient) *WorkerRunner
```

NewWorkerRunnerWithApiClient creates a new WorkerRunner which uses the provided client.APIClient to communicate with Conductor.

<a name="WorkerRunner.DecreaseBatchSize"></a>
### func \(\*WorkerRunner\) [DecreaseBatchSize](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L164>)

```go
func (c *WorkerRunner) DecreaseBatchSize(taskName string, batchSize int) error
```

DecreaseBatchSize decreases the batch size used for all workers running the provided task.

<a name="WorkerRunner.GetBatchSizeForAll"></a>
### func \(\*WorkerRunner\) [GetBatchSizeForAll](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L514>)

```go
func (c *WorkerRunner) GetBatchSizeForAll() (batchSizeByTaskName map[string]int)
```

GetBatchSizeForAll returns a map from taskName to batch size for all batch sizes currently registered with this WorkerRunner.

<a name="WorkerRunner.GetBatchSizeForTask"></a>
### func \(\*WorkerRunner\) [GetBatchSizeForTask](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L525>)

```go
func (c *WorkerRunner) GetBatchSizeForTask(taskName string) (batchSize int)
```

GetBatchSizeForTask retrieves the current batch size for the provided task.

<a name="WorkerRunner.GetPollIntervalForTask"></a>
### func \(\*WorkerRunner\) [GetPollIntervalForTask](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L502>)

```go
func (c *WorkerRunner) GetPollIntervalForTask(taskName string) (pollInterval time.Duration, err error)
```

GetPollIntervalForTask retrieves the poll interval for all tasks running the provided taskName. An error is returned if no pollInterval has been registered for the provided task.

<a name="WorkerRunner.IncreaseBatchSize"></a>
### func \(\*WorkerRunner\) [IncreaseBatchSize](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L141>)

```go
func (c *WorkerRunner) IncreaseBatchSize(taskName string, batchSize int) error
```

IncreaseBatchSize increases the batch size used for all workers running the provided task.

<a name="WorkerRunner.Pause"></a>
### func \(\*WorkerRunner\) [Pause](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L190>)

```go
func (c *WorkerRunner) Pause(taskName string)
```

Pause pauses all workers running the provided task. When paused, workers will not poll for new tasks and no new goroutines are started. However it does not stop any goroutines running. Workers must be resumed at a later time using Resume. Failing to call \`Resume\(\)\` on a WorkerRunner running one or more workers can result in a goroutine leak.

<a name="WorkerRunner.Resume"></a>
### func \(\*WorkerRunner\) [Resume](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L198>)

```go
func (c *WorkerRunner) Resume(taskName string)
```

Resume all running workers for the provided taskName. If workers for the provided task are not paused, calling this method has no impact.

<a name="WorkerRunner.SetBatchSize"></a>
### func \(\*WorkerRunner\) [SetBatchSize](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L116>)

```go
func (c *WorkerRunner) SetBatchSize(taskName string, batchSize int) error
```

SetBatchSize can be used to set the batch size for all workers running the provided task.

<a name="WorkerRunner.SetPollIntervalForTask"></a>
### func \(\*WorkerRunner\) [SetPollIntervalForTask](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L492>)

```go
func (c *WorkerRunner) SetPollIntervalForTask(taskName string, pollInterval time.Duration) error
```

SetPollIntervalForTask sets the pollInterval for all workers running the task with the provided taskName.

<a name="WorkerRunner.SetSleepOnGenericError"></a>
### func \(\*WorkerRunner\) [SetSleepOnGenericError](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L93>)

```go
func (c *WorkerRunner) SetSleepOnGenericError(duration time.Duration)
```

SetSleepOnGenericError Sets the time for which to wait before continuing to poll/execute when there is an error Default is 200 millis, and this function can be used to increase/decrease the duration of the wait time Useful to avoid excessive logs in the worker when there are intermittent issues

<a name="WorkerRunner.StartWorker"></a>
### func \(\*WorkerRunner\) [StartWorker](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L111>)

```go
func (c *WorkerRunner) StartWorker(taskName string, executeFunction model.WorkerTaskFunction, batchSize int, pollInterval time.Duration) error
```

StartWorker starts a worker on a new goroutine, which polls conductor periodically for tasks matching the provided taskName and, if any are available, uses executeFunction to run them on a separate goroutine. Each call to StartWorker starts a new goroutine which performs batch polling to retrieve as many tasks from Conductor as are available, up to the batchSize set for the task. This func additionally sets the pollInterval and increases the batch size for the task, which applies to all tasks shared by this WorkerRunner with the same taskName.

<a name="WorkerRunner.StartWorkerWithDomain"></a>
### func \(\*WorkerRunner\) [StartWorkerWithDomain](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L101>)

```go
func (c *WorkerRunner) StartWorkerWithDomain(taskName string, executeFunction model.WorkerTaskFunction, batchSize int, pollInterval time.Duration, domain string) error
```

StartWorkerWithDomain starts a polling worker on a new goroutine, which only polls for tasks using the provided domain. Equivalent to:

```
StartWorkerWithDomain(taskName, executeFunction, batchSize, pollInterval, "")
```

<a name="WorkerRunner.WaitWorkers"></a>
### func \(\*WorkerRunner\) [WaitWorkers](<https://github.com/vkantchev/conductor-client-golang/blob/main/sdk/worker/worker_runner.go#L212>)

```go
func (c *WorkerRunner) WaitWorkers()
```

WaitWorkers uses an internal waitgroup to block the calling thread until all workers started by this WorkerRunner have been stopped.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
