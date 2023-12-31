# Authoring Workflows with the Go SDK

## A simple two-step workflow

```go

//API client instance with server address
apiClient := client.NewAPIClient(settings.NewHttpSettings("http://localhost:8080/api"))

//Create new workflow manager
manager := manager.NewWorkflowManager(apiClient)

//Create a new WorkflowDefEx instance
conductorWorkflow := workflow.NewWorkflowDefEx(manager).
    Name("my_first_workflow").
    Version(1).
    OwnerEmail("hello@swiftsoftwaregroup.com")

//now, let's add a couple of simple tasks
conductorWorkflow.
	Add(workflow.NewSimpleTask("simple_task_2", "simple_task_1")).
    Add(workflow.NewSimpleTask("simple_task_1", "simple_task_2"))

//Register the workflow with server
conductorWorkflow.Register(true)        //Overwrite the existing definition with the new one
```
### Execute Workflow

#### Using Workflow Executor to start previously registered workflow

```go
//Input can be either a map or a struct that is serializable to a JSON map
workflowInput := map[string]interface{}{}

workflowId, err := manager.StartWorkflow(&model.StartWorkflowRequest{
    Name:  conductorWorkflow.GetName(),
    Input: workflowInput,
})
```

#### Using Workflow Executor to synchronously execute a workflow and get the output as a result

```go
//Input can be either a map or a struct that is serializable to a JSON map
workflowInput := map[string]interface{}{}

workflowRun, err := manager.RunWorkflow(&model.StartWorkflowRequest{Name: wf.GetName(), Version: &version, Input: workflowInput}, "")
//workfowRun is a struct that contains the output of the workflow execution
type WorkflowRun struct {
    CorrelationId string                 `json:"correlationId,omitempty"`
    CreateTime    int64                  `json:"createTime,omitempty"`
    CreatedBy     string                 `json:"createdBy,omitempty"`
    Input         map[string]interface{} `json:"input,omitempty"`
    Output        map[string]interface{} `json:"output,omitempty"`
    Priority      int32                  `json:"priority,omitempty"`
    RequestId     string                 `json:"requestId,omitempty"`
    Status        string                 `json:"status,omitempty"`
    Tasks         []Task                 `json:"tasks,omitempty"`
    UpdateTime    int64                  `json:"updateTime,omitempty"`
    Variables     map[string]interface{} `json:"variables,omitempty"`
    WorkflowId    string                 `json:"workflowId,omitempty"`
}
```
**Note:** Synchronous workflow execution is useful for workflows that complete in few seconds at max.  For longer running workflows use `StartWorkflow` and use the Id of the workflow to monitor the output.

#### Using struct instance as workflow input
```go
type WorkflowInput struct {
    Name string
    Address []string
}
//...
workflowId, err := manager.StartWorkflow(&model.StartWorkflowRequest{
  Name:  conductorWorkflow.GetName(),
  Input: &WorkflowInput{
  Name: "John Doe",
  Address: []string{"street", "city", "zip"},
  },
})
```
### Workflow Management APIs

See [Docs](docs/workflow.md) for APIs to start, pause, resume, terminate, search and get workflow execution status.


