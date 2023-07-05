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

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	v1 "github.com/nitrictech/nitric/core/pkg/api/nitric/v1"
	"github.com/nitrictech/nitric/core/pkg/worker/adapter"
)

// Represents a locally running http server
type HttpWorker struct {
	port int

	adapter.Adapter
}

var _ Worker = &HttpWorker{}

func (*HttpWorker) HandlesTrigger(req *v1.TriggerRequest) bool {
	// Can handle any given HTTP request
	return req.GetHttp() != nil
}

// TODO: We should proxy the request instead as best we can
// This will allow this worker to be added to a pool and used generically
// however we will want to make sure we don't miss anything in context
func (h *HttpWorker) HandleTrigger(ctx context.Context, req *v1.TriggerRequest) (*v1.TriggerResponse, error) {
	if req.GetHttp() == nil {
		return nil, fmt.Errorf("http worker cannot handle Event requests")
	}

	targetHost, err := url.Parse(fmt.Sprintf("http://localhost:%d", h.port))
	if err != nil {
		return nil, err
	}

	newHeader := http.Header{}

	targetPath := targetHost.JoinPath(req.GetHttp().Path)
	httpReq, err := http.NewRequest(req.GetHttp().GetMethod(), targetPath.String(), io.NopCloser(bytes.NewReader(req.Data)))
	if err != nil {
		return nil, err
	}

	for k, v := range req.GetHttp().Headers {
		for _, val := range v.Value {
			// Replace forwarded authorization with base authorization so the user gets the expected headers
			if k == "X-Forwarded-Authorization" && newHeader["Authorization"] == nil {
				k = "Authorization"
			}

			httpReq.Header.Add(k, val)
		}
	}

	// forward the request to the server
	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	responseHeaders := map[string]*v1.HeaderValue{}

	for k, v := range res.Header {
		responseHeaders[k] = &v1.HeaderValue{
			Value: v,
		}
	}

	return &v1.TriggerResponse{
		Data: body,
		Context: &v1.TriggerResponse_Http{
			Http: &v1.HttpResponseContext{
				Status:  int32(res.StatusCode),
				Headers: responseHeaders,
			},
		},
	}, nil
}

func (h *HttpWorker) GetPort() int {
	return h.port
}

func NewHttpWorker(adapter adapter.Adapter, port int) *HttpWorker {
	return &HttpWorker{
		Adapter: adapter,
		port:    port,
	}
}
