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

func TestSimpleWorkerHost(t *testing.T) {
	httpSettings := settings.NewHttpDefaultSettings()

	workerHost := worker.NewWorkerHost(httpSettings)
	if workerHost == nil {
		t.Fail()
	}
}

func TestWorkerHost(t *testing.T) {
	httpSettings := settings.NewHttpDefaultSettings()
	apiClient := client.NewAPIClient(httpSettings)

	workerHost := worker.NewWorkerHostWithApiClient(apiClient)

	if workerHost == nil {
		t.Fail()
	}
}

func TestPauseResume(t *testing.T) {
	httpSettings := settings.NewHttpDefaultSettings()
	apiClient := client.NewAPIClient(httpSettings)

	workerHost := worker.NewWorkerHostWithApiClient(apiClient)

	workerHost.StartWorker("test", TaskWorker, 21, time.Second)
	workerHost.Pause("test")
	assert.Equal(t, 21, workerHost.GetBatchSizeForTask("test"))

	workerHost.Resume("test")
	assert.Equal(t, 21, workerHost.GetBatchSizeForTask("test"))
}

func TaskWorker(task *model.WorkerTask) (interface{}, error) {
	return map[string]interface{}{
		"zip": "10121",
	}, nil
}
