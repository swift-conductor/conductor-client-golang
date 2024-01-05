//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package workflow

import (
	"encoding/json"

	"swiftconductor.com/swift-conductor-client/sdk/model"

	log "github.com/sirupsen/logrus"
)

type TimeoutPolicy string

const (
	TimeOutWorkflow TimeoutPolicy = "TIME_OUT_WF"
	AlertOnly       TimeoutPolicy = "ALERT_ONLY"
)

type WorkflowDefEx struct {
	manager                       *WorkflowManager
	name                          string
	version                       int32
	description                   string
	ownerEmail                    string
	tasks                         []WorkflowTaskInterface
	timeoutPolicy                 TimeoutPolicy
	timeoutSeconds                int64
	failureWorkflow               string
	inputParameters               []string
	outputParameters              map[string]interface{}
	inputTemplate                 map[string]interface{}
	variables                     map[string]interface{}
	restartable                   bool
	workflowStatusListenerEnabled bool
}

func NewWorkflowDefEx(manager *WorkflowManager) *WorkflowDefEx {
	return &WorkflowDefEx{
		manager:       manager,
		timeoutPolicy: AlertOnly,
		restartable:   true,
	}
}

func (workflow *WorkflowDefEx) Name(name string) *WorkflowDefEx {
	workflow.name = name
	return workflow
}

func (workflow *WorkflowDefEx) Version(version int32) *WorkflowDefEx {
	workflow.version = version
	return workflow
}

func (workflow *WorkflowDefEx) Description(description string) *WorkflowDefEx {
	workflow.description = description
	return workflow
}

func (workflow *WorkflowDefEx) TimeoutPolicy(timeoutPolicy TimeoutPolicy, timeoutSeconds int64) *WorkflowDefEx {
	workflow.timeoutPolicy = timeoutPolicy
	workflow.timeoutSeconds = timeoutSeconds
	return workflow
}

func (workflow *WorkflowDefEx) TimeoutSeconds(timeoutSeconds int64) *WorkflowDefEx {
	workflow.timeoutSeconds = timeoutSeconds
	return workflow
}

// FailureWorkflow name of the workflow to execute when this workflow fails.
// Failure workflows can be used for handling compensation logic
func (workflow *WorkflowDefEx) FailureWorkflow(failureWorkflow string) *WorkflowDefEx {
	workflow.failureWorkflow = failureWorkflow
	return workflow
}

// Restartable if the workflow can be restarted after it has reached terminal state.
// Set this to false if restarting workflow can have side effects
func (workflow *WorkflowDefEx) Restartable(restartable bool) *WorkflowDefEx {
	workflow.restartable = restartable
	return workflow
}

// WorkflowStatusListenerEnabled if the workflow status listener need to be enabled.
func (workflow *WorkflowDefEx) WorkflowStatusListenerEnabled(workflowStatusListenerEnabled bool) *WorkflowDefEx {
	workflow.workflowStatusListenerEnabled = workflowStatusListenerEnabled
	return workflow
}

// OutputParameters Workflow outputs. Workflow output follows similar structure as task inputs
// See https://swiftconductor.com/devguide/how-tos/Tasks/task-inputs.html for more details
func (workflow *WorkflowDefEx) OutputParameters(outputParameters interface{}) *WorkflowDefEx {
	workflow.outputParameters = getInputAsMap(outputParameters)
	return workflow
}

// InputTemplate template input to the workflow.  Can have combination of variables (e.g. ${workflow.input.abc}) and
// static values
func (workflow *WorkflowDefEx) InputTemplate(inputTemplate interface{}) *WorkflowDefEx {
	workflow.inputTemplate = getInputAsMap(inputTemplate)
	return workflow
}

// Variables Workflow variables are set using SET_VARIABLE task.  Excellent way to maintain business state
// e.g. Variables can maintain business/user specific states which can be queried and inspected to find out the state of the workflow
func (workflow *WorkflowDefEx) Variables(variables interface{}) *WorkflowDefEx {
	workflow.variables = getInputAsMap(variables)
	return workflow
}

// InputParameters List of the input parameters to the workflow.  Used ONLY for the documentation purpose.
func (workflow *WorkflowDefEx) InputParameters(inputParameters ...string) *WorkflowDefEx {
	workflow.inputParameters = inputParameters
	return workflow
}

func (workflow *WorkflowDefEx) OwnerEmail(ownerEmail string) *WorkflowDefEx {
	workflow.ownerEmail = ownerEmail
	return workflow
}

func (workflow *WorkflowDefEx) GetName() (name string) {
	return workflow.name
}

func (workflow *WorkflowDefEx) GetOutputParameters() (outputParameters map[string]interface{}) {
	return workflow.outputParameters
}

func (workflow *WorkflowDefEx) GetVersion() (version int32) {
	return workflow.version
}

func (workflow *WorkflowDefEx) Add(task WorkflowTaskInterface) *WorkflowDefEx {
	workflow.tasks = append(workflow.tasks, task)
	return workflow
}

// Register the workflow definition with the server. If overwrite is set, the definition on the server will be overwritten.
// When not set, the call fails if there is any change in the workflow definition between the server and what is being registered.
func (workflow *WorkflowDefEx) Register(overwrite bool) error {
	return workflow.manager.RegisterWorkflow(overwrite, workflow.ToWorkflowDef())
}

// Register the workflow definition with the server. If overwrite is set, the definition on the server will be overwritten.
// When not set, the call fails if there is any change in the workflow definition between the server and what is being registered.
func (workflow *WorkflowDefEx) UnRegister() error {
	return workflow.manager.UnRegisterWorkflow(workflow.name, workflow.version)
}

// StartWorkflowWithInput RunWorkflowWithInput Execute the workflow with specific input.  The input struct MUST be serializable to JSON
// Returns the workflow Id that can be used to monitor and get the status of the workflow execution
func (workflow *WorkflowDefEx) StartWorkflowWithInput(input interface{}) (workflowId string, err error) {
	version := workflow.GetVersion()
	return workflow.manager.StartWorkflow(
		&model.StartWorkflowRequest{
			Name:        workflow.GetName(),
			Version:     version,
			Input:       getInputAsMap(input),
			WorkflowDef: workflow.ToWorkflowDef(),
		},
	)
}

// StartWorkflow starts the workflow execution with startWorkflowRequest that allows you to specify more details like task domains, correlationId etc.
// Returns the ID of the newly created workflow
func (workflow *WorkflowDefEx) StartWorkflow(startWorkflowRequest *model.StartWorkflowRequest) (workflowId string, err error) {
	startWorkflowRequest.WorkflowDef = workflow.ToWorkflowDef()
	return workflow.manager.StartWorkflow(startWorkflowRequest)
}

// RunWorkflowWithInput Execute the workflow with specific input and wait for the workflow to complete or until the task specified as waitUntil is completed.
// waitUntilTask Reference name of the task which MUST be completed before returning the output.  if specified as empty string, then the call waits until the
// workflow completes or reaches the timeout (as specified on the server)
// The input struct MUST be serializable to JSON
// Returns the workflow output
func (workflow *WorkflowDefEx) RunWorkflowWithInput(input interface{}, waitUntilTask string) (worfklowRun *model.WorkflowRun, err error) {
	version := workflow.GetVersion()
	return workflow.manager.RunWorkflow(
		&model.StartWorkflowRequest{
			Name:        workflow.GetName(),
			Version:     version,
			Input:       getInputAsMap(input),
			WorkflowDef: workflow.ToWorkflowDef(),
		},
		waitUntilTask,
	)
}

// StartWorkflowsAndMonitorExecution Starts the workflow execution and returns a channel that can be used to monitor the workflow execution
// This method is useful for short duration workflows that are expected to complete in few seconds.  For long-running workflows use GetStatus APIs to periodically check the status
func (workflow *WorkflowDefEx) StartWorkflowsAndMonitorExecution(startWorkflowRequest *model.StartWorkflowRequest) (runningChannel RunningWorkflowChannel, err error) {
	workflowId, err := workflow.StartWorkflow(startWorkflowRequest)
	if err != nil {
		return nil, err
	}

	return workflow.manager.MonitorExecution(workflowId)
}

func getInputAsMap(input interface{}) map[string]interface{} {
	if input == nil {
		return nil
	}
	casted, ok := input.(map[string]interface{})
	if ok {
		return casted
	}

	data, err := json.Marshal(input)
	if err != nil {
		log.Debug(
			"Failed to parse input",
			", reason: ", err.Error(),
		)
		return nil
	}

	var parsedInput map[string]interface{}
	json.Unmarshal(data, &parsedInput)
	return parsedInput
}

// ToWorkflowDef converts the workflow to the JSON serializable format
func (workflow *WorkflowDefEx) ToWorkflowDef() *model.WorkflowDef {
	return &model.WorkflowDef{
		Name:                          workflow.name,
		Description:                   workflow.description,
		Version:                       workflow.version,
		Tasks:                         getWorkflowTasksFromConductorWorkflow(workflow),
		InputParameters:               workflow.inputParameters,
		OutputParameters:              workflow.outputParameters,
		FailureWorkflow:               workflow.failureWorkflow,
		SchemaVersion:                 2,
		OwnerEmail:                    workflow.ownerEmail,
		TimeoutPolicy:                 string(workflow.timeoutPolicy),
		TimeoutSeconds:                workflow.timeoutSeconds,
		Variables:                     workflow.variables,
		InputTemplate:                 workflow.inputTemplate,
		Restartable:                   workflow.restartable,
		WorkflowStatusListenerEnabled: workflow.workflowStatusListenerEnabled,
	}
}

func getWorkflowTasksFromConductorWorkflow(workflow *WorkflowDefEx) []model.WorkflowTask {
	workflowTasks := make([]model.WorkflowTask, 0)
	for _, task := range workflow.tasks {
		workflowTasks = append(
			workflowTasks,
			task.toWorkflowTask()...,
		)
	}
	return workflowTasks
}
