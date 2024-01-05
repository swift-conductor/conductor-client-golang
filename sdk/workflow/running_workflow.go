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
	"fmt"
	"time"

	"github.com/swift-conductor/conductor-client-golang/sdk/model"
)

type RunningWorkflowChannel chan *model.Workflow

type RunningWorkflow struct {
	WorkflowId             string
	RunningWorkflowChannel RunningWorkflowChannel
	Err                    error
	CompletedWorkflow      *model.Workflow
}

func NewRunningWorkflow(workflowId string, runningWorkflowChannel RunningWorkflowChannel, err error) *RunningWorkflow {
	return &RunningWorkflow{
		WorkflowId:             workflowId,
		RunningWorkflowChannel: runningWorkflowChannel,
		Err:                    err,
		CompletedWorkflow:      nil,
	}
}

func (rw *RunningWorkflow) WaitForCompletionUntilTimeout(timeout time.Duration) (workflow *model.Workflow, err error) {
	select {
	case workflow, ok := <-rw.RunningWorkflowChannel:
		if !ok {
			return nil, fmt.Errorf("channel closed")
		}
		rw.CompletedWorkflow = workflow
		return workflow, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout")
	}
}
