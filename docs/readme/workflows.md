# Authoring Workflows with the Go SDK

## A simple two-step workflow

```go
// API client instance with server address
httpSettings := settings.NewHttpSettings("http://localhost:8080/api")
apiClient := client.NewAPIClient(httpSettings)

// create a new WorkflowBuilder instance
builder := workflow.NewWorkflowBuilder().
    Name("my_first_workflow").
    Version(1).
    OwnerEmail("hello@swiftsoftwaregroup.com").
    // add a couple of custom tasks
    Add(workflow.NewCustomTask("custom_task", "custom_task_1")).
    Add(workflow.NewCustomTask("custom_task", "custom_task_2"))

// create new workflow manager
manager := manager.NewWorkflowManager(apiClient)

// Register the workflow with server
workflowDef := builder.ToWorkflowDef()
manager.RegisterWorkflow(workflowDef)
```

### Execute Workflow

```go
// Input can be either a map or a struct that is serializable to a JSON map
workflowInput := map[string]interface{}{}

startWorkflowRequest := model.StartWorkflowRequest{
    Name:    workflowDef.Name,
    Version: workflowDef.Version,
}

workflowId, err := manager.StartWorkflow(startWorkflowRequest)
```

#### Wait for a workflow to finish

```go
runningChannel, err := manager.MonitorExecution(workflowId)

_, err := workflow.WaitForWorkflowCompletionUntilTimeout(runningChannel, 20*time.Second)
```

#### Using struct instance as workflow input

```go
type WorkflowInput struct {
    Name string
    Address []string
}

//...

workflowId, err := manager.StartWorkflow(&model.StartWorkflowRequest{
    Name:  workflowDef.Name,
    Input: &WorkflowInput{
        Name: "John Doe",
        Address: []string{"street", "city", "zip"},
    },
})
```

