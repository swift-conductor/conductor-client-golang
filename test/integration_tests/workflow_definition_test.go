package integration_tests

import (
	"context"
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
	workflow := testdata.NewKitchenSinkWorkflowDefEx(testdata.WorkflowManager)
	err := workflow.Register(true)
	if err != nil {
		t.Fatalf("Failed to register workflow: %s, reason: %s", workflow.GetName(), err.Error())
	}

	startWorkers()

	run, err := executeWorkflowWithRetries(workflow, map[string]interface{}{
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
	wf := workflow.NewWorkflowDefEx(manager)
	wf.Name("temp_wf_" + strconv.Itoa(time.Now().Nanosecond())).Version(1)
	wf = wf.Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))
	err := wf.Register(true)

	assert.NoError(t, err, "Failed to register workflow")

	id, err := manager.StartWorkflow(&model.StartWorkflowRequest{Name: wf.GetName()})
	assert.NoError(t, err, "Failed to start workflow")

	execution, err := manager.GetWorkflow(id, true)
	assert.NoError(t, err, "Failed to get workflow execution")
	assert.Equal(t, model.CompletedWorkflow, execution.Status, "Workflow is not in the completed state")

	err = manager.RemoveWorkflow(id)
	assert.NoError(t, err, "Failed to remove workflow execution")

	_, err = manager.GetWorkflow(id, true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such workflow by Id")

	_, err = testdata.MetadataClient.UnregisterWorkflowDef(
		context.Background(),
		wf.GetName(),
		wf.GetVersion(),
	)
	assert.NoError(t, err, "Failed to delete workflow definition ", err)
}

func TestRunWorkflow(t *testing.T) {
	manager := testdata.WorkflowManager
	wf := workflow.NewWorkflowDefEx(manager).
		Name("temp_wf_2_" + strconv.Itoa(time.Now().Nanosecond())).
		Version(1).
		OwnerEmail("hello@swiftsoftwaregroup.com")
	wf = wf.Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))
	wf.OutputParameters(map[string]interface{}{
		"param1": "Test",
		"param2": 123,
	})
	err := wf.Register(true)

	assert.NoError(t, err, "Failed to register workflow")
	version := wf.GetVersion()
	run, err := executeWorkflowWithRetriesWithStartWorkflowRequest(
		&model.StartWorkflowRequest{
			Name:    wf.GetName(),
			Version: version,
		},
	)
	assert.NoError(t, err, "Failed to start workflow")
	assert.Equal(t, string(model.CompletedWorkflow), run.Status)

	execution, err := manager.GetWorkflow(run.WorkflowId, true)
	assert.NoError(t, err, "Failed to get workflow execution")
	assert.Equal(t, model.CompletedWorkflow, execution.Status, "Workflow is not in the completed state")

	_, err = testdata.MetadataClient.UnregisterWorkflowDef(
		context.Background(),
		wf.GetName(),
		wf.GetVersion(),
	)
	assert.NoError(t, err, "Failed to delete workflow definition ", err)
}

func TestRunWorkflowWithCorrelationIds(t *testing.T) {
	manager := testdata.WorkflowManager
	correlationId1 := "correlationId1-" + uuid.New().String()
	correlationId2 := "correlationId2-" + uuid.New().String()

	httpTaskWorkflow1 := workflow.NewWorkflowDefEx(testdata.WorkflowManager).
		Name("TEST_GO_WORKFLOW_HTTP" + correlationId1).
		OwnerEmail("test@test.com").
		Version(1).
		Add(common.TestHttpTask)
	httpTaskWorkflow2 := workflow.NewWorkflowDefEx(testdata.WorkflowManager).
		Name("TEST_GO_WORKFLOW_HTTP" + correlationId2).
		OwnerEmail("test@test.com").
		Version(1).
		Add(common.TestHttpTask)
	_, err := httpTaskWorkflow1.StartWorkflow(&model.StartWorkflowRequest{CorrelationId: correlationId1})
	if err != nil {
		t.Fatal(err)
	}
	_, err = httpTaskWorkflow2.StartWorkflow(&model.StartWorkflowRequest{CorrelationId: correlationId2})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)
	workflows, err := manager.GetByCorrelationIdsAndNames(true, true,
		[]string{correlationId1, correlationId2}, []string{httpTaskWorkflow1.GetName(), httpTaskWorkflow2.GetName()})
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
	wf := workflow.NewWorkflowDefEx(manager).
		Name("TEST_GO_SET_VAR_USED_AS_FAILURE").
		Version(1).
		OwnerEmail("test@test.com").
		Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))
	err := testdata.ValidateWorkflowRegistration(wf)
	if err != nil {
		t.Fatal(err)
	}

	workflowWait := workflow.NewWorkflowDefEx(testdata.WorkflowManager).
		Name("TEST_GO_WORKFLOW_WAIT_CONDUCTOR").
		Version(1).
		OwnerEmail("test@test.com").
		Add(workflow.NewWaitTask("termination_wait")).
		FailureWorkflow(wf.GetName())
	err = testdata.ValidateWorkflowRegistration(workflowWait)
	if err != nil {
		t.Fatal(err)
	}

	id, err := workflowWait.StartWorkflow(&model.StartWorkflowRequest{})
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
	wf := workflow.NewWorkflowDefEx(manager).
		Name("temp_wf_3_" + strconv.Itoa(time.Now().Nanosecond())).
		Version(1).
		OwnerEmail("test@test.com")
	wf = wf.Add(workflow.NewSetVariableTask("set_var").Input("var_value", 42))
	wf.OutputParameters(map[string]interface{}{
		"param1": "Test",
		"param2": 123,
	})
	err := wf.Register(true)

	assert.NoError(t, err, "Failed to register workflow")
	run, err := executeWorkflowWithRetries(wf, map[string]interface{}{
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

	_, err = testdata.MetadataClient.UnregisterWorkflowDef(
		context.Background(),
		wf.GetName(),
		wf.GetVersion(),
	)
	assert.NoError(t, err, "Failed to delete workflow definition ", err)
}

func startWorkers() {
	testdata.WorkerRunner.StartWorker("simple_task", testdata.SimpleWorker, 10, 100*time.Millisecond)
	testdata.WorkerRunner.StartWorker("dynamic_fork_prep", testdata.DynamicForkWorker, 3, 100*time.Millisecond)
}

func executeWorkflowWithRetries(wf *workflow.WorkflowDefEx, workflowInput interface{}) (*model.WorkflowRun, error) {
	for attempt := 0; attempt < retryLimit; attempt += 1 {
		workflowRun, err := wf.RunWorkflowWithInput(workflowInput, "")
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
	for attempt := 1; attempt <= retryLimit; attempt += 1 {
		workflowRun, err := testdata.WorkflowManager.RunWorkflow(startWorkflowRequest, "")
		if err != nil {
			time.Sleep(time.Duration(attempt+2) * time.Second)
			fmt.Printf("Failed to execute workflow, reason: %s", err.Error())
			continue
		}
		return workflowRun, nil
	}
	return nil, fmt.Errorf("exhausted retries for workflow execution")
}
