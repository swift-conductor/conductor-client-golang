package integration_tests

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/stretchr/testify/assert"
	"github.com/swift-conductor/conductor-client-golang/internal/testdata"
	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/workflow"
	"github.com/swift-conductor/conductor-client-golang/test/common"
)

const retryLimit = 5

func TestWorkflowCreation(t *testing.T) {
	// create and register custom task
	customTaskDef := model.TaskDef{Name: "custom_task", OwnerEmail: "test@test.com"}
	err := testdata.RegisterTasks(customTaskDef)
	if err != nil {
		t.Fatal(err)
	}

	// create and register workflow
	builder := testdata.NewKitchenSinkWorkflowBuilder()

	workflowDef := builder.ToWorkflowDef()
	err = testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatalf("Failed to register workflow: %s, reason: %s", builder.GetName(), err.Error())
	}

	// start workers for custom task and dynamic fork
	startWorkers()

	workflowId, err := runWorkflowWithInput(workflowDef, map[string]interface{}{
		"key1": "input1",
		"key2": 101,
	})
	if err != nil {
		t.Fatalf("Failed to complete the workflow, reason: %s", err)
	}

	time.Sleep(10 * time.Second)

	manager := testdata.WorkflowManager

	run, err := manager.GetWorkflow(workflowId, false)
	assert.NoError(t, err)
	assert.NotEmpty(t, run, "Workflow is null", run)

	assert.Equal(t, model.CompletedWorkflow, run.Status)
	assert.Equal(t, "input1", run.Input["key1"])
}

func TestRemoveWorkflow(t *testing.T) {
	manager := testdata.WorkflowManager

	builder := workflow.NewWorkflowBuilder().Name("temp_wf_" + strconv.Itoa(time.Now().Nanosecond()))
	builder.Version(1).OwnerEmail("test@test.com")
	builder.Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))

	workflowDef := builder.ToWorkflowDef()

	err := manager.RegisterWorkflow(workflowDef)
	assert.NoError(t, err, "Failed to register workflow")

	id, err := manager.StartWorkflow(&model.StartWorkflowRequest{Name: workflowDef.Name, WorkflowDef: workflowDef})
	assert.NoError(t, err, "Failed to start workflow")

	execution, err := manager.GetWorkflow(id, true)
	assert.NoError(t, err, "Failed to get workflow execution")
	assert.Equal(t, model.CompletedWorkflow, execution.Status, "Workflow is not in the completed state")

	err = manager.RemoveWorkflow(id)
	assert.NoError(t, err, "Failed to remove workflow execution")

	_, err = manager.GetWorkflow(id, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such workflow by Id")

	err = manager.UnRegisterWorkflow(workflowDef.Name, workflowDef.Version)
	assert.NoError(t, err, "Failed to delete workflow definition ", err)
}

func TestRunWorkflow(t *testing.T) {
	manager := testdata.WorkflowManager

	builder := workflow.NewWorkflowBuilder().
		Name("temp_wf_2_" + strconv.Itoa(time.Now().Nanosecond())).
		Version(1).
		OwnerEmail("test@test.com")

	builder = builder.Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))

	builder.OutputParameters(map[string]interface{}{
		"param1": "Test",
		"param2": 123,
	})

	workflowDef := builder.ToWorkflowDef()

	err := manager.RegisterWorkflow(workflowDef)
	assert.NoError(t, err, "Failed to register workflow")

	startWorkflowRequest := model.StartWorkflowRequest{
		Name:    builder.GetName(),
		Version: builder.GetVersion(),
	}

	workflowId, err := runWorkflowWithStartWorkflowRequest(&startWorkflowRequest)
	assert.NoError(t, err, "Failed to start workflow")

	run, err := manager.GetWorkflow(workflowId, false)
	assert.NoError(t, err, "Failed to get workflow execution")
	assert.Equal(t, model.CompletedWorkflow, run.Status, "Workflow is not in the completed state")

	err = manager.UnRegisterWorkflow(workflowDef.Name, workflowDef.Version)
	assert.NoError(t, err, "Failed to delete workflow definition ", err)
}

func TestRunWorkflowWithCorrelationIds(t *testing.T) {
	manager := testdata.WorkflowManager

	correlationId1 := "correlationId1-" + uuid.New().String()
	correlationId2 := "correlationId2-" + uuid.New().String()

	builder_1 := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_HTTP_" + correlationId1).
		OwnerEmail("test@test.com").
		Version(1).
		Add(common.TestHttpTask)

	builder_2 := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_HTTP_" + correlationId2).
		OwnerEmail("test@test.com").
		Version(1).
		Add(common.TestHttpTask)

	workflowDef_1 := builder_1.ToWorkflowDef()
	workflowDef_2 := builder_2.ToWorkflowDef()

	_, err := manager.StartWorkflow(&model.StartWorkflowRequest{CorrelationId: correlationId1, WorkflowDef: workflowDef_1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.StartWorkflow(&model.StartWorkflowRequest{CorrelationId: correlationId2, WorkflowDef: workflowDef_2})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(3 * time.Second)

	workflows, err := manager.GetByCorrelationIds(workflowDef_1.Name, true, true, correlationId1)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, workflows, correlationId1)
	assert.NotEmpty(t, workflows[correlationId1])
	assert.Equal(t, workflows[correlationId1][0].CorrelationId, correlationId1)

	workflows, err = manager.GetByCorrelationIds(workflowDef_2.Name, true, true, correlationId2)
	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, workflows, correlationId2)
	assert.NotEmpty(t, workflows[correlationId2])
	assert.Equal(t, workflows[correlationId2][0].CorrelationId, correlationId2)
}

func TestTerminateWorkflowWithFailure(t *testing.T) {

	manager := testdata.WorkflowManager
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_SET_VAR_USED_AS_FAILURE").
		Version(1).
		OwnerEmail("test@test.com").
		Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))

	wfDef := builder.ToWorkflowDef()
	err := testdata.RegisterWorkflow(wfDef)
	if err != nil {
		t.Fatal(err)
	}

	builderTerm := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_WAIT_TO_BE_TERMINATED").
		Version(1).
		OwnerEmail("test@test.com").
		FailureWorkflow(wfDef.Name).
		WorkflowStatusListenerEnabled(true).
		Add(workflow.NewWaitTask("termination_wait"))

	termWfDef := builderTerm.ToWorkflowDef()
	err = testdata.RegisterWorkflow(termWfDef)
	if err != nil {
		t.Fatal(err)
	}

	termWfId, err := manager.StartWorkflow(&model.StartWorkflowRequest{WorkflowDef: termWfDef})
	if err != nil {
		t.Fatal(err)
	}

	err = manager.TerminateWithFailure(termWfId, "Terminated to trigger failure workflow", true)
	if err != nil {
		t.Fatal(err)
	}

	terminatedWfRun, err := manager.GetWorkflow(termWfId, false)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, terminatedWfRun.ReasonForIncompletion, "Terminated to trigger failure workflow")
	assert.NotEmpty(t, terminatedWfRun.Output["conductor.failure_workflow"])
}

func TestRunWorkflowSync(t *testing.T) {
	manager := testdata.WorkflowManager

	builder := workflow.NewWorkflowBuilder().
		Name("temp_wf_3_" + strconv.Itoa(time.Now().Nanosecond())).
		Version(1).
		OwnerEmail("test@test.com")

	builder = builder.Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))

	builder.OutputParameters(map[string]interface{}{
		"param1": "Test",
		"param2": 123,
	})

	workflowDef := builder.ToWorkflowDef()

	err := manager.RegisterWorkflow(workflowDef)
	assert.NoError(t, err, "Failed to register workflow")

	workflowId, err := runWorkflowWithInput(workflowDef, map[string]interface{}{
		"key1": "input1",
		"key2": 101,
	})
	if err != nil {
		t.Fatalf("Failed to complete the workflow, reason: %s", err)
	}

	run, err := manager.GetWorkflow(workflowId, false)
	assert.NoError(t, err, "Failed to get workflow execution")
	assert.Equal(t, model.CompletedWorkflow, run.Status, "Workflow is not in the completed state")

	err = manager.UnRegisterWorkflow(workflowDef.Name, workflowDef.Version)
	assert.NoError(t, err, "Failed to delete workflow definition ", err)
}

func startWorkers() {
	testdata.WorkerHost.StartWorker("custom_task", testdata.CustomWorker, 10, 100*time.Millisecond)
	testdata.WorkerHost.StartWorker("dynamic_fork_prep", testdata.DynamicForkWorker, 3, 100*time.Millisecond)
}

func runWorkflowWithInput(workflowDef *model.WorkflowDef, workflowInput interface{}) (string, error) {
	manager := testdata.WorkflowManager

	for attempt := 0; attempt < retryLimit; attempt += 1 {
		workflowId, err := manager.StartWorkflowWithInput(workflowDef, workflowInput)
		if err != nil {
			time.Sleep(time.Duration(attempt+2) * time.Second)
			fmt.Println("Failed to execute workflow, reason: " + err.Error())
			continue
		}

		workflowExecutionChannel, err := manager.MonitorExecution(workflowId)
		if err != nil {
			fmt.Println("Failed to execute workflow, reason: " + err.Error())
		}

		_, err = workflow.WaitForWorkflowCompletionUntilTimeout(workflowExecutionChannel, 20*time.Second)
		if err != nil {
			fmt.Println("Wait for completion error: " + err.Error())
		}

		return workflowId, nil
	}

	return "", fmt.Errorf("exhausted retries for workflow execution")
}

func runWorkflowWithStartWorkflowRequest(startWorkflowRequest *model.StartWorkflowRequest) (string, error) {
	manager := testdata.WorkflowManager
	for attempt := 1; attempt <= retryLimit; attempt += 1 {
		workflowId, err := manager.StartWorkflow(startWorkflowRequest)
		if err != nil {
			time.Sleep(time.Duration(attempt+2) * time.Second)
			fmt.Printf("Failed to execute workflow, reason: %s", err.Error())
			continue
		}

		manager.MonitorExecution(workflowId)
		return workflowId, nil
	}

	return "", fmt.Errorf("exhausted retries for workflow execution")
}
