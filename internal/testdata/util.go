//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package testdata

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/swift-conductor/conductor-client-golang/sdk/client"
	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/settings"
	"github.com/swift-conductor/conductor-client-golang/sdk/worker"
	"github.com/swift-conductor/conductor-client-golang/sdk/workflow"

	log "github.com/sirupsen/logrus"
)

const (
	BASE_URL = "CONDUCTOR_SERVER_URL"
)

var (
	apiClient = getApiClient()
)

var (
	MetadataClient = client.MetadataResourceApiService{
		APIClient: apiClient,
	}
	TaskClient = client.TaskResourceApiService{
		APIClient: apiClient,
	}
	WorkflowClient = client.WorkflowResourceApiService{
		APIClient: apiClient,
	}
	EventClient = client.EventResourceApiService{
		APIClient: apiClient,
	}
)

var WorkflowManager = workflow.NewWorkflowManager(apiClient)
var WorkerHost = worker.NewWorkerHostWithApiClient(apiClient)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.ErrorLevel)
}

func ValidateWorkflowDaemon(waitTime time.Duration, outputChannel chan error, workflowId string, expectedOutput map[string]interface{}, expectedStatus model.WorkflowStatus) {
	time.Sleep(waitTime)
	workflow, _, err := WorkflowClient.GetExecutionStatus(
		context.Background(),
		workflowId,
		nil,
	)
	if err != nil {
		outputChannel <- err
		return
	}
	if workflow.Status != expectedStatus {
		outputChannel <- fmt.Errorf(
			"workflow status different than expected, workflowId: %s, workflowStatus: %s",
			workflow.WorkflowId, workflow.Status,
		)
		return
	}
	if !reflect.DeepEqual(workflow.Output, expectedOutput) {
		outputChannel <- fmt.Errorf(
			"workflow output is different than expected, workflowId: %s, output: %+v",
			workflow.WorkflowId, workflow.Output,
		)
		return
	}
	outputChannel <- nil
}

func getApiClient() *client.APIClient {
	return client.NewAPIClient(
		getHttpSettings(),
	)
}

func getHttpSettings() *settings.HttpSettings {
	var base_url = os.Getenv(BASE_URL)

	if base_url != "" {
		return settings.NewHttpSettings(base_url)
	}

	return settings.NewHttpSettings("http://localhost:8080/api")
}

func StartWorkflows(workflowQty int, workflowName string) ([]string, error) {
	workflowIdList := make([]string, workflowQty)
	for i := 0; i < workflowQty; i += 1 {
		workflowId, _, err := WorkflowClient.StartWorkflow(
			context.Background(),
			make(map[string]interface{}),
			workflowName,
			nil,
		)
		if err != nil {
			return nil, err
		}
		log.Debug(
			"Started workflow",
			", workflowName: ", workflowName,
			", workflowId: ", workflowId,
		)
		workflowIdList[i] = workflowId
	}
	return workflowIdList, nil
}

func RunWorkflow(workflowDef *model.WorkflowDef, timeout time.Duration, expectedStatus model.WorkflowStatus) error {
	err := RegisterWorkflow(workflowDef)
	if err != nil {
		return err
	}

	workflowId, err := WorkflowManager.StartWorkflowWithInput(workflowDef, make(map[string]interface{}))
	if err != nil {
		return err
	}
	log.Debug("Started workflowId: ", workflowId)

	workflowExecutionChannel, err := WorkflowManager.MonitorExecution(workflowId)
	if err != nil {
		return err
	}
	log.Debug("Generated workflowExecutionChannel for workflowId: ", workflowId)

	workflow, err := workflow.WaitForWorkflowCompletionUntilTimeout(workflowExecutionChannel, timeout)
	if err != nil {
		return err
	}
	log.Debug("Workflow completed, workflowId: ", workflowId)

	if workflow.Status != expectedStatus {
		return fmt.Errorf("workflow finished with unexpected status: %s", workflow.Status)
	}

	return nil
}

func RunWorkflowsBulk(workflowDef *model.WorkflowDef, timeout time.Duration, count int) error {
	err := RegisterWorkflow(workflowDef)
	if err != nil {
		return err
	}

	startWorkflowRequests := make([]*model.StartWorkflowRequest, count)
	for i := 0; i < count; i += 1 {
		startWorkflowRequests[i] = model.NewStartWorkflowRequest(
			workflowDef.Name,
			workflowDef.Version,
			"",
			make(map[string]interface{}),
		)
	}

	runningWorkflows := WorkflowManager.StartWorkflows(true, startWorkflowRequests...)

	WorkflowManager.WaitForRunningWorkflowsUntilTimeout(timeout, runningWorkflows...)
	for _, runningWorkflow := range runningWorkflows {
		if runningWorkflow.Err != nil {
			return err
		}

		if runningWorkflow.CompletedWorkflow == nil {
			return fmt.Errorf("invalid completed workflows")
		}

		if runningWorkflow.CompletedWorkflow.Status != model.CompletedWorkflow {
			return fmt.Errorf("workflow finished with status: %s", runningWorkflow.CompletedWorkflow.Status)
		}
	}

	return nil
}

func RegisterTasks(taskDefs ...model.TaskDef) error {
	response, err := MetadataClient.RegisterTaskDef(
		context.Background(),
		taskDefs,
	)
	if err != nil {
		log.Debug(
			"Failed to validate task registration. Reason: ", err.Error(),
			", response: ", *response,
		)
		return err
	}
	return nil
}

func RegisterWorkflow(workflowDef *model.WorkflowDef) error {
	existinWorkflowDef, _, err := MetadataClient.Get(context.Background(), workflowDef.Name, nil)
	if err == nil {
		MetadataClient.UnregisterWorkflowDef(context.Background(), existinWorkflowDef.Name, existinWorkflowDef.Version)
	}

	for attempt := 0; attempt < 5; attempt += 1 {
		err := WorkflowManager.RegisterWorkflow(workflowDef)
		if err != nil {
			time.Sleep(time.Duration(attempt+2) * time.Second)
			fmt.Println("Failed to validate workflow registration, reason: " + err.Error())
			continue
		}

		return nil
	}

	return fmt.Errorf("exhausted retries")
}

func UnregisterWorkflow(workflowDef *model.WorkflowDef) error {
	for attempt := 0; attempt < 5; attempt += 1 {
		err := WorkflowManager.UnRegisterWorkflow(workflowDef.Name, workflowDef.Version)
		if err != nil {
			time.Sleep(time.Duration(attempt+2) * time.Second)
			fmt.Println("Failed to validate workflow deletion, reason: " + err.Error())
			continue
		}

		return nil
	}

	return fmt.Errorf("exhausted retries")
}
