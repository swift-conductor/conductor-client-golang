//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package integration_tests

import (
	"context"
	"fmt"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/swift-conductor/conductor-client-golang/internal/testdata"
	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/workflow"
	"github.com/swift-conductor/conductor-client-golang/test/common"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.ErrorLevel)
}

func TestHttpTask(t *testing.T) {
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_HTTP").
		Version(1).
		OwnerEmail("test@test.com").
		WorkflowStatusListenerEnabled(true).
		Add(common.TestHttpTask)

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RunWorkflow(workflowDef, common.WorkflowValidationTimeout, model.CompletedWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.RunWorkflowsBulk(workflowDef, common.WorkflowValidationTimeout, common.WorkflowBulkCount)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func CustomTask(t *testing.T) {
	err := testdata.RegisterTasks(*common.TestCustomTask.ToTaskDef())
	if err != nil {
		t.Fatal(err)
	}

	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_CUSTOM").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestCustomTask)

	err = testdata.WorkerHost.StartWorker(
		common.TestCustomTask.ReferenceName(),
		testdata.CustomWorker,
		testdata.WorkerCount,
		testdata.WorkerPollInterval,
	)
	if err != nil {
		t.Fatal(err)
	}

	workflowDef := builder.ToWorkflowDef()

	err = testdata.RunWorkflow(workflowDef, common.WorkflowValidationTimeout, model.CompletedWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.RunWorkflowsBulk(workflowDef, common.WorkflowValidationTimeout, common.WorkflowBulkCount)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.WorkerHost.DecreaseBatchSize(
		common.TestCustomTask.ReferenceName(),
		testdata.WorkerCount,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func CustomTaskWithoutRetryCount(t *testing.T) {
	taskToRegister := common.TestCustomTask.ToTaskDef()
	taskToRegister.RetryCount = 0
	err := testdata.RegisterTasks(*taskToRegister)
	if err != nil {
		t.Fatal(err)
	}

	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_CUSTOM").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestCustomTask)

	err = testdata.WorkerHost.StartWorker(
		common.TestCustomTask.ReferenceName(),
		testdata.CustomWorker,
		testdata.WorkerCount,
		testdata.WorkerPollInterval,
	)
	if err != nil {
		t.Fatal(err)
	}

	workflowDef := builder.ToWorkflowDef()

	err = testdata.RunWorkflow(workflowDef, common.WorkflowValidationTimeout, model.CompletedWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.RunWorkflowsBulk(workflowDef, common.WorkflowValidationTimeout, common.WorkflowBulkCount)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.WorkerHost.DecreaseBatchSize(
		common.TestCustomTask.ReferenceName(),
		testdata.WorkerCount,
	)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func TestInlineTask(t *testing.T) {
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_INLINE_TASK").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestInlineTask)

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RunWorkflow(workflowDef, common.WorkflowValidationTimeout, model.CompletedWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.RunWorkflowsBulk(workflowDef, common.WorkflowValidationTimeout, common.WorkflowBulkCount)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func TestSqsEventTask(t *testing.T) {
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_EVENT_SQS").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestSqsEventTask)

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func TestConductorEventTask(t *testing.T) {
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_EVENT_CONDUCTOR").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestConductorEventTask)

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func TestKafkaPublishTask(t *testing.T) {
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_KAFKA_PUBLISH").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestKafkaPublishTask)

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func TestDoWhileTask(t *testing.T) {
}

func TestTerminateTask(t *testing.T) {
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_TERMINATE").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestTerminateTask)

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func TestSwitchTask(t *testing.T) {
	builder := workflow.NewWorkflowBuilder().
		Name("TEST_GO_WORKFLOW_SWITCH").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestSwitchTask)

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func TestDynamicForkWorkflow(t *testing.T) {
	builder := workflow.NewWorkflowBuilder().
		Name("dynamic_workflow_array_sub_workflow").
		Version(1).
		OwnerEmail("test@test.com").
		Add(createDynamicForkTask())

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal()
	}

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func createDynamicForkTask() *workflow.DynamicForkTask {
	return workflow.NewDynamicForkTaskWithoutPrepareTask(
		"dynamic_workflow_array_sub_workflow",
	).Input(
		"forkTaskWorkflow", "extract_user",
	).Input(
		"forkTaskInputs", []map[string]interface{}{
			{
				"input": "value1",
			},
			{
				"sub_workflow_2_inputs": map[string]interface{}{
					"key":  "value",
					"key2": 23,
				},
			},
		},
	)
}

func TestComplexSwitchWorkflow(t *testing.T) {
	builder := testdata.GetWorkflowBuilderWithComplexSwitchTask()

	workflowDef := builder.ToWorkflowDef()

	err := testdata.RegisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(err)
	}

	receivedWf, _, err := testdata.MetadataClient.Get(context.Background(), workflowDef.Name, nil)
	if err != nil {
		t.Fatal(err)
	}

	counter := countMultipleSwitchInnerTasks(receivedWf.Tasks...)
	assert.Equal(t, 7, counter)

	err = testdata.UnregisterWorkflow(workflowDef)
	if err != nil {
		t.Fatal(
			"Failed to delete workflow. Reason: ", err.Error(),
		)
	}
}

func countMultipleSwitchInnerTasks(tasks ...model.WorkflowTask) int {
	counter := 0
	for _, task := range tasks {
		counter += countSwitchInnerTasks(task)
	}
	return counter
}

func countSwitchInnerTasks(task model.WorkflowTask) int {
	fmt.Println(task.Type_)
	counter := 1
	if task.Type_ != "SWITCH" {
		return counter
	}
	for _, value := range task.DecisionCases {
		counter += countMultipleSwitchInnerTasks(value...)
	}
	return counter
}
