// Copyright 2021 Nitric Technologies Pty Ltd.
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

package worker

type WrappedWorkerFn func(Worker) Worker

type InstrumentedWorkerPool struct {
	WorkerPool
	Wrapper WrappedWorkerFn
}

var _ WorkerPool = &InstrumentedWorkerPool{}

func (iwp *InstrumentedWorkerPool) GetWorker(opts *GetWorkerOptions) (Worker, error) {
	w, err := iwp.WorkerPool.GetWorker(opts)
	if err != nil {
		return w, err
	}

	return iwp.Wrapper(w), nil
}
