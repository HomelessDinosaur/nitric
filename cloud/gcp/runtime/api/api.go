// Copyright 2021 Nitric Technologies Pty Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"

	"github.com/nitrictech/nitric/cloud/gcp/runtime/resource"
	apispb "github.com/nitrictech/nitric/core/pkg/proto/apis/v1"
	"github.com/nitrictech/nitric/core/pkg/workers/apis"
)

type GcpApiGatewayProvider struct {
	provider *resource.GcpResourceService
	*apis.RouteWorkerManager
}

var _ apispb.ApiServer = &GcpApiGatewayProvider{}

func (g *GcpApiGatewayProvider) Details(ctx context.Context, req *apispb.ApiDetailsRequest) (*apispb.ApiDetailsResponse, error) {
	gwDetails, err := g.provider.GetApiGatewayDetails(ctx, req.ApiName)
	if err != nil {
		return nil, err
	}

	return &apispb.ApiDetailsResponse{
		Url: gwDetails.Url,
	}, nil
}

func NewGcpApiGatewayProvider(provider *resource.GcpResourceService) *GcpApiGatewayProvider {
	return &GcpApiGatewayProvider{
		provider:           provider,
		RouteWorkerManager: apis.New(),
	}
}
