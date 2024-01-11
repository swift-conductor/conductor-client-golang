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
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/swift-conductor/conductor-client-golang/internal/testdata"
	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/workflow"
	"github.com/swift-conductor/conductor-client-golang/test/common"
)

func TestWorkerBatchSize(t *testing.T) {
	customTaskWorkflow := workflow.NewWorkflowBuilder(testdata.WorkflowManager).
		Name("TEST_GO_WORKFLOW_CUSTOM").
		Version(1).
		OwnerEmail("test@test.com").
		Add(common.TestCustomTask)

	err := testdata.WorkerHost.StartWorker(
		common.TestCustomTask.ReferenceName(),
		testdata.CustomWorker,
		5,
		testdata.WorkerPollInterval,
	)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	if testdata.WorkerHost.GetBatchSizeForTask(common.TestCustomTask.ReferenceName()) != 5 {
		t.Fatal("unexpected batch size")
	}

	err = testdata.ValidateWorkflowBulk(customTaskWorkflow, common.WorkflowValidationTimeout, common.WorkflowBulkQty)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.WorkerHost.SetBatchSize(
		common.TestCustomTask.ReferenceName(),
		0,
	)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	if testdata.WorkerHost.GetBatchSizeForTask(common.TestCustomTask.ReferenceName()) != 0 {
		t.Fatal("unexpected batch size")
	}

	err = testdata.WorkerHost.SetBatchSize(
		common.TestCustomTask.ReferenceName(),
		8,
	)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	if testdata.WorkerHost.GetBatchSizeForTask(common.TestCustomTask.ReferenceName()) != 8 {
		t.Fatal("unexpected batch size")
	}

	err = testdata.ValidateWorkflowBulk(customTaskWorkflow, common.WorkflowValidationTimeout, common.WorkflowBulkQty)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFaultyWorker(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	taskName := "TEST_GO_FAULTY_TASK"
	wf := workflow.NewWorkflowBuilder(testdata.WorkflowManager).
		Name("TEST_GO_FAULTY_WORKFLOW").
		Version(1).
		OwnerEmail("test@test.com").
		Add(workflow.NewCustomTask(taskName, taskName))
	err := wf.Register(true)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.WorkerHost.StartWorker(
		taskName,
		testdata.FaultyWorker,
		5,
		testdata.WorkerPollInterval,
	)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.ValidateWorkflow(wf, 5*time.Second, model.FailedWorkflow)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWorkerWithNonRetryableError(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)

	taskName := "TEST_GO_NON_RETRYABLE_ERROR_TASK"

	wf := workflow.NewWorkflowBuilder(testdata.WorkflowManager).
		Name("TEST_GO_NON_RETRYABLE_ERROR_WF").
		Version(1).
		OwnerEmail("test@test.com").
		Add(workflow.NewCustomTask(taskName, taskName))

	err := wf.Register(true)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.WorkerHost.StartWorker(taskName, testdata.FaultyWorker, 5, testdata.WorkerPollInterval)
	if err != nil {
		t.Fatal(err)
	}

	err = testdata.ValidateWorkflow(wf, 5*time.Second, model.FailedWorkflow)
	if err != nil {
		t.Fatal(err)
	}
}
