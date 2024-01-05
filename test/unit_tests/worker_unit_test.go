//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with
//  the License. You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
//  an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
//  specific language governing permissions and limitations under the License.

package unit_tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/swift-conductor/conductor-client-golang/sdk/client"
	"github.com/swift-conductor/conductor-client-golang/sdk/model"
	"github.com/swift-conductor/conductor-client-golang/sdk/settings"
	"github.com/swift-conductor/conductor-client-golang/sdk/worker"
)

func TestSimpleWorkerRunner(t *testing.T) {
	taskRunner := worker.NewWorkerRunner(settings.NewHttpDefaultSettings())
	if taskRunner == nil {
		t.Fail()
	}
}

func TestWorkerRunner(t *testing.T) {
	apiClient := client.NewAPIClient(
		settings.NewHttpDefaultSettings(),
	)

	taskRunner := worker.NewWorkerRunnerWithApiClient(
		apiClient,
	)

	if taskRunner == nil {
		t.Fail()
	}
}

func TestPauseResume(t *testing.T) {
	apiClient := client.NewAPIClient(
		settings.NewHttpDefaultSettings(),
	)
	taskRunner := worker.NewWorkerRunnerWithApiClient(
		apiClient,
	)
	taskRunner.StartWorker("test", TaskWorker, 21, time.Second)
	taskRunner.Pause("test")
	assert.Equal(t, 21, taskRunner.GetBatchSizeForTask("test"))
	taskRunner.Resume("test")
	assert.Equal(t, 21, taskRunner.GetBatchSizeForTask("test"))

}

func TaskWorker(task *model.WorkerTask) (interface{}, error) {
	return map[string]interface{}{
		"zip": "10121",
	}, nil
}
