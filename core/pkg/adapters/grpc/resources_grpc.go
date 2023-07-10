package grpc

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

import (
	"context"
	"fmt"

	v1 "github.com/nitrictech/nitric/core/pkg/api/nitric/v1"
	"github.com/nitrictech/nitric/core/pkg/plugins/resource"
)

type ResourcesServiceServer struct {
	v1.UnimplementedResourceServiceServer
	plugin resource.ResourceService
}

type ResourceServiceOption = func(*ResourcesServiceServer)

func WithResourcePlugin(plugin resource.ResourceService) ResourceServiceOption {
	return func(srv *ResourcesServiceServer) {
		if plugin != nil {
			srv.plugin = plugin
		}
	}
}

func (rs *ResourcesServiceServer) checkPluginRegistered() error {
	if rs.plugin == nil {
		return NewPluginNotRegisteredError("Resource")
	}

	return nil
}

func (rs *ResourcesServiceServer) Declare(ctx context.Context, req *v1.ResourceDeclareRequest) (*v1.ResourceDeclareResponse, error) {
	if err := rs.checkPluginRegistered(); err != nil {
		return nil, err
	}

	err := rs.plugin.Declare(ctx, req)
	if err != nil {
		return nil, err
	}

	return &v1.ResourceDeclareResponse{}, nil
}

func populateDetails(details *resource.DetailsResponse[any], response *v1.ResourceDetailsResponse) error {
	switch det := details.Detail.(type) {
	case resource.ApiDetails:
		response.Details = &v1.ResourceDetailsResponse_Api{
			Api: &v1.ApiResourceDetails{
				Url: det.URL,
			},
		}
		return nil
	case resource.WebsocketDetails:
		response.Details = &v1.ResourceDetailsResponse_Websocket{
			Websocket: &v1.WebsocketResourceDetails{
				Url: det.URL,
			},
		}
		return nil
	default:
		return fmt.Errorf("unsupported details type")
	}
}

var resourceTypeMap = map[v1.ResourceType]resource.ResourceType{
	v1.ResourceType_Api:       resource.ResourceType_Api,
	v1.ResourceType_Websocket: resource.ResourceType_Websocket,
}

func (rs *ResourcesServiceServer) Details(ctx context.Context, req *v1.ResourceDetailsRequest) (*v1.ResourceDetailsResponse, error) {
	if err := rs.checkPluginRegistered(); err != nil {
		return nil, err
	}

	cType, ok := resourceTypeMap[req.Resource.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", req.Resource.Type)
	}

	d, err := rs.plugin.Details(ctx, cType, req.Resource.Name)
	if err != nil {
		return nil, err
	}

	resp := &v1.ResourceDetailsResponse{
		Id:       d.Id,
		Provider: d.Provider,
		Service:  d.Service,
	}

	err = populateDetails(d, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func NewResourcesServiceServer(opts ...ResourceServiceOption) v1.ResourceServiceServer {
	// Default server implementation
	srv := &ResourcesServiceServer{
		plugin: &resource.UnimplementResourceService{},
	}

	// Apply options
	for _, o := range opts {
		o(srv)
	}

	return srv
}
