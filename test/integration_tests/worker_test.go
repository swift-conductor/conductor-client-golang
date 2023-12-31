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

	"github.com/swift-conductor/conductor-client-golang/internal/testdata"
	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/workflow"
	"github.com/swift-conductor/conductor-client-golang/test/common"
	"github.com/sirupsen/logrus"
)

func TestWorkerBatchSize(t *testing.T) {
	simpleTaskWorkflow := workflow.NewConductorWorkflow(testdata.WorkflowExecutor).
		Name("TEST_GO_WORKFLOW_SIMPLE").
		Version(1).
		Add(common.TestSimpleTask)
	err := testdata.TaskRunner.StartWorker(
		common.TestSimpleTask.ReferenceName(),
		testdata.SimpleWorker,
		5,
		testdata.WorkerPollInterval,
	)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if testdata.TaskRunner.GetBatchSizeForTask(common.TestSimpleTask.ReferenceName()) != 5 {
		t.Fatal("unexpected batch size")
	}
	err = testdata.ValidateWorkflowBulk(simpleTaskWorkflow, common.WorkflowValidationTimeout, common.WorkflowBulkQty)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.TaskRunner.SetBatchSize(
		common.TestSimpleTask.ReferenceName(),
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if testdata.TaskRunner.GetBatchSizeForTask(common.TestSimpleTask.ReferenceName()) != 0 {
		t.Fatal("unexpected batch size")
	}
	err = testdata.TaskRunner.SetBatchSize(
		common.TestSimpleTask.ReferenceName(),
		8,
	)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if testdata.TaskRunner.GetBatchSizeForTask(common.TestSimpleTask.ReferenceName()) != 8 {
		t.Fatal("unexpected batch size")
	}
	err = testdata.ValidateWorkflowBulk(simpleTaskWorkflow, common.WorkflowValidationTimeout, common.WorkflowBulkQty)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFaultyWorker(t *testing.T) {
	logrus.SetLevel(logrus.ErrorLevel)
	taskName := "TEST_GO_FAULTY_TASK"
	wf := workflow.NewConductorWorkflow(testdata.WorkflowExecutor).
		Name("TEST_GO_FAULTY_WORKFLOW").
		Version(1).
		Add(workflow.NewSimpleTask(taskName, taskName))
	err := wf.Register(true)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.TaskRunner.StartWorker(
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
	wf := workflow.NewConductorWorkflow(testdata.WorkflowExecutor).
		Name("TEST_GO_NON_RETRYABLE_ERROR_WF").
		Version(1).
		Add(workflow.NewSimpleTask(taskName, taskName))
	err := wf.Register(true)
	if err != nil {
		t.Fatal(err)
	}
	err = testdata.TaskRunner.StartWorker(
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
