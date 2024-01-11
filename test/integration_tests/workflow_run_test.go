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
	builder := testdata.NewKitchenSinkWorkflowBuilder()

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatalf("Failed to register workflow: %s, reason: %s", builder.GetName(), err.Error())
	}

	startWorkers()

	run, err := executeWorkflowWithRetries(workflowDef, map[string]interface{}{
		"key1": "input1",
		"key2": 101,
	})

	if err != nil {
		t.Fatalf("Failed to complete the workflow, reason: %s", err)
	}

	assert.NotEmpty(t, run, "Workflow is null", run)
	assert.Equal(t, string(model.CompletedWorkflow), run.Status)
	assert.Equal(t, "input1", run.Input["key1"])
}

func TestRemoveWorkflow(t *testing.T) {
	manager := testdata.WorkflowManager

	builder := workflow.NewWorkflowBuilder()
	builder.Name("temp_wf_" + strconv.Itoa(time.Now().Nanosecond())).Version(1)
	builder = builder.Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))

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

	run, err := executeWorkflowWithRetriesWithStartWorkflowRequest(&startWorkflowRequest)
	assert.NoError(t, err, "Failed to start workflow")
	assert.Equal(t, string(model.CompletedWorkflow), run.Status)

	execution, err := manager.GetWorkflow(run.WorkflowId, true)
	assert.NoError(t, err, "Failed to get workflow execution")
	assert.Equal(t, model.CompletedWorkflow, execution.Status, "Workflow is not in the completed state")

	err = manager.UnRegisterWorkflow(workflowDef.Name, workflowDef.Version)
	assert.NoError(t, err, "Failed to delete workflow definition ", err)
}

func TestRunWorkflowWithCorrelationIds(t *testing.T) {
	manager := testdata.WorkflowManager

	correlationId1 := "correlationId1-" + uuid.New().String()
	correlationId2 := "correlationId2-" + uuid.New().String()

	builder_1 := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_HTTP" + correlationId1).
		OwnerEmail("test@test.com").
		Version(1).
		Add(common.TestHttpTask)

	builder_2 := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_HTTP" + correlationId2).
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

	workflows, err := manager.GetByCorrelationIdsAndNames(true, true,
		[]string{correlationId1, correlationId2}, []string{builder_1.GetName(), builder_2.GetName()})

	if err != nil {
		t.Fatal(err)
	}

	assert.Contains(t, workflows, correlationId1)
	assert.Contains(t, workflows, correlationId2)
	assert.NotEmpty(t, workflows[correlationId1])
	assert.NotEmpty(t, workflows[correlationId2])
	assert.Equal(t, workflows[correlationId1][0].CorrelationId, correlationId1)
	assert.Equal(t, workflows[correlationId2][0].CorrelationId, correlationId2)
}

func TestTerminateWorkflowWithFailure(t *testing.T) {

	manager := testdata.WorkflowManager
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_SET_VAR_USED_AS_FAILURE").
		Version(1).
		OwnerEmail("test@test.com").
		Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(err)
	}

	builderWait := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_WAIT_CONDUCTOR").
		Version(1).
		OwnerEmail("test@test.com").
		Add(workflow.NewWaitTask("termination_wait")).
		FailureWorkflow(workflowDef.Name)

	workflowDefWait := builderWait.ToWorkflowDef()

	err = testdata.RegisterWorkflow(workflowDefWait)
	if err != nil {
		t.Fatal(err)
	}

	id, err := manager.StartWorkflow(&model.StartWorkflowRequest{})
	if err != nil {
		t.Fatal(err)
	}

	err = manager.TerminateWithFailure(id, "Terminated to trigger failure workflow", true)
	if err != nil {
		t.Fatal(err)
	}

	terminatedWfStatus, err := manager.GetWorkflow(id, false)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEmpty(t, terminatedWfStatus.Output["conductor.failure_workflow"])
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

	run, err := executeWorkflowWithRetries(workflowDef, map[string]interface{}{
		"key1": "input1",
		"key2": 101,
	})
	if err != nil {
		t.Fatalf("Failed to complete the workflow, reason: %s", err)
	}
	assert.NotEmpty(t, run, "Workflow is null", run)
	assert.Equal(t, string(model.CompletedWorkflow), run.Status)

	execution, err := manager.GetWorkflow(run.WorkflowId, true)
	assert.NoError(t, err, "Failed to get workflow execution")
	assert.Equal(t, model.CompletedWorkflow, execution.Status, "Workflow is not in the completed state")

	err = manager.UnRegisterWorkflow(workflowDef.Name, workflowDef.Version)
	assert.NoError(t, err, "Failed to delete workflow definition ", err)
}

func startWorkers() {
	testdata.WorkerHost.StartWorker("custom_task", testdata.CustomWorker, 10, 100*time.Millisecond)
	testdata.WorkerHost.StartWorker("dynamic_fork_prep", testdata.DynamicForkWorker, 3, 100*time.Millisecond)
}

func executeWorkflowWithRetries(workflowDef *model.WorkflowDef, workflowInput interface{}) (*model.WorkflowRun, error) {
	manager := testdata.WorkflowManager

	for attempt := 0; attempt < retryLimit; attempt += 1 {

		workflowRun, err := manager.RunWorkflowWithInput(workflowDef, "", "")
		if err != nil {
			time.Sleep(time.Duration(attempt+2) * time.Second)
			fmt.Println("Failed to execute workflow, reason: " + err.Error())
			continue
		}
		return workflowRun, nil
	}
	return nil, fmt.Errorf("exhausted retries for workflow execution")
}

func executeWorkflowWithRetriesWithStartWorkflowRequest(startWorkflowRequest *model.StartWorkflowRequest) (*model.WorkflowRun, error) {
	manager := testdata.WorkflowManager
	for attempt := 1; attempt <= retryLimit; attempt += 1 {
		workflowRun, err := manager.RunWorkflow(startWorkflowRequest, "")
		if err != nil {
			time.Sleep(time.Duration(attempt+2) * time.Second)
			fmt.Printf("Failed to execute workflow, reason: %s", err.Error())
			continue
		}

		return workflowRun, nil
	}

	return nil, fmt.Errorf("exhausted retries for workflow execution")
}
