// Copyright 2021 Nitric Pty Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package queue_service_test

import (
	"encoding/json"

	mocks "github.com/nitric-dev/membrane/mocks/dev_storage"
	"github.com/nitric-dev/membrane/sdk"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	queue_plugin "github.com/nitric-dev/membrane/plugins/queue/dev"
)

var _ = Describe("Queue", func() {
	Context("SendBatch", func() {
		When("The queue is empty", func() {
			mockStorageDriver := mocks.NewMockStorageDriver(&mocks.MockStorageDriverOptions{})
			queuePlugin, _ := queue_plugin.NewWithStorageDriver(mockStorageDriver)
			task := sdk.NitricTask{
				ID:          "1234",
				PayloadType: "test-payload",
				Payload: map[string]interface{}{
					"Test": "Test",
				},
			}
			tasks := []sdk.NitricTask{task}
			taskBytes, _ := json.Marshal(tasks)
			It("Should store the events in the queue", func() {
				resp, err := queuePlugin.SendBatch("test", tasks)
				By("Not returning an error")
				Expect(err).ShouldNot(HaveOccurred())

				By("Returning No failed messages")
				Expect(resp.FailedTasks).To(BeEmpty())

				By("Storing the sent message, in the given queue")
				Expect(mockStorageDriver.GetStoredItems()["/nitric/queues/test"]).ToNot(BeNil())

				By("Storing the content of the given message")
				Expect(mockStorageDriver.GetStoredItems()["/nitric/queues/test"]).To(BeEquivalentTo(taskBytes))
			})
		})

		When("The queue is not empty", func() {
			task := sdk.NitricTask{
				ID:          "1234",
				PayloadType: "test-payload",
				Payload: map[string]interface{}{
					"Test": "Test",
				},
			}
			tasks := []sdk.NitricTask{task}
			evtsBytes, _ := json.Marshal(tasks)
			mockStorageDriver := mocks.NewMockStorageDriver(&mocks.MockStorageDriverOptions{
				StoredItems: map[string][]byte{
					"/nitric/queues/test": evtsBytes,
				},
			})
			queuePlugin, _ := queue_plugin.NewWithStorageDriver(mockStorageDriver)

			It("Should append to the existing queue", func() {
				resp, err := queuePlugin.SendBatch("test", tasks)
				By("Not returning an error")
				Expect(err).ShouldNot(HaveOccurred())

				By("Having no Failed Messages")
				Expect(resp.FailedTasks).To(BeEmpty())

				By("Storing the sent message, in the given queue")
				Expect(mockStorageDriver.GetStoredItems()["/nitric/queues/test"]).ToNot(BeNil())

				var messages []sdk.NitricEvent
				bytes := mockStorageDriver.GetStoredItems()["/nitric/queues/test"]
				json.Unmarshal(bytes, &messages)
				By("Having 2 messages on the Queue")
				Expect(messages).To(HaveLen(2))
			})
		})
	})

	Context("Receive", func() {
		When("The queue is empty", func() {
			tasksBytes, _ := json.Marshal([]sdk.NitricTask{})
			mockStorageDriver := mocks.NewMockStorageDriver(&mocks.MockStorageDriverOptions{
				StoredItems: map[string][]byte{
					"/nitric/queues/test": tasksBytes,
				},
			})
			queuePlugin, _ := queue_plugin.NewWithStorageDriver(mockStorageDriver)

			It("Should return an empty slice of queue items", func() {
				depth := uint32(10)
				items, err := queuePlugin.Receive(sdk.ReceiveOptions{
					QueueName: "test",
					Depth:     &depth,
				})
				By("Not returning an error")
				Expect(err).ShouldNot(HaveOccurred())

				By("Returning an empty slice")
				Expect(items).To(HaveLen(0))
			})
		})

		When("The queue is not empty", func() {
			task := sdk.NitricEvent{
				ID:          "1234",
				PayloadType: "test-payload",
				Payload: map[string]interface{}{
					"Test": "Test",
				},
			}
			tasks := []sdk.NitricEvent{task}
			taskBytes, _ := json.Marshal(tasks)
			mockStorageDriver := mocks.NewMockStorageDriver(&mocks.MockStorageDriverOptions{
				StoredItems: map[string][]byte{
					"/nitric/queues/test": taskBytes,
				},
			})
			queuePlugin, _ := queue_plugin.NewWithStorageDriver(mockStorageDriver)

			It("Should append to the existing queue", func() {
				depth := uint32(10)
				items, err := queuePlugin.Receive(sdk.ReceiveOptions{
					QueueName: "test",
					Depth:     &depth,
				})
				By("Not returning an error")
				Expect(err).ShouldNot(HaveOccurred())

				By("Returning 1 item")
				Expect(items).To(HaveLen(1))

				var messages []sdk.NitricTask
				bytes := mockStorageDriver.GetStoredItems()["/nitric/queues/test"]
				json.Unmarshal(bytes, &messages)
				By("Having no remaining messages on the Queue")
				Expect(messages).To(HaveLen(0))
			})
		})

		When("The queue depth is 15", func() {
			task := sdk.NitricTask{
				ID:          "1234",
				PayloadType: "test-payload",
				Payload: map[string]interface{}{
					"Test": "Test",
				},
			}
			tasks := []sdk.NitricTask{}

			// Add 15 items to the queue
			for i := 0; i < 15; i++ {
				tasks = append(tasks, task)
			}

			taskBytes, _ := json.Marshal(tasks)
			mockStorageDriver := mocks.NewMockStorageDriver(&mocks.MockStorageDriverOptions{
				StoredItems: map[string][]byte{
					"/nitric/queues/test": taskBytes,
				},
			})
			queuePlugin, _ := queue_plugin.NewWithStorageDriver(mockStorageDriver)

			When("Requested depth is 10", func() {
				It("Should return 10 items", func() {
					depth := uint32(10)
					items, err := queuePlugin.Receive(sdk.ReceiveOptions{
						QueueName: "test",
						Depth:     &depth,
					})
					By("Not returning an error")
					Expect(err).ShouldNot(HaveOccurred())

					By("Returning 10 item")
					Expect(items).To(HaveLen(10))

					var messages []sdk.NitricTask
					bytes := mockStorageDriver.GetStoredItems()["/nitric/queues/test"]
					json.Unmarshal(bytes, &messages)
					By("Having 5 remaining messages on the Queue")
					Expect(messages).To(HaveLen(5))
				})
			})
		})
	})

	Context("Complete", func() {
		// Currently the local queue complete method is a stub that always returns successfully.
		// We may consider adding more realistic behavior if that is useful in future.
		When("it always returns successfully", func() {
			mockStorageDriver := mocks.NewMockStorageDriver(&mocks.MockStorageDriverOptions{})
			queuePlugin, _ := queue_plugin.NewWithStorageDriver(mockStorageDriver)

			It("Should retnot return an error", func() {
				err := queuePlugin.Complete("test-queue", "test-id")
				By("Not returning an error")
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
