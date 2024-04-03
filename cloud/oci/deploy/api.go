package deploy

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/nitrictech/nitric/cloud/common/deploy/resources"
	"github.com/nitrictech/nitric/cloud/common/deploy/tags"
	deploymentspb "github.com/nitrictech/nitric/core/pkg/proto/deployments/v1"
	"github.com/pulumi/pulumi-oci/sdk/go/oci/apigateway"
	"github.com/pulumi/pulumi-oci/sdk/go/oci/functions"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Api struct {
	pulumi.ResourceState
	Name string
}

type nameUrlPair struct {
	name      string
	invokeUrl string
}

func (n *NitricOCIPulumiProvider) Api(ctx *pulumi.Context, parent pulumi.Resource, name string, config *deploymentspb.Api) error {
	var err error

	opts := []pulumi.ResourceOption{pulumi.Parent(parent)}

	if config.GetOpenapi() == "" {
		return fmt.Errorf("oci provider can only deploy OpenAPI specs")
	}

	openapiDoc := &openapi3.T{}
	err = openapiDoc.UnmarshalJSON([]byte(config.GetOpenapi()))
	if err != nil {
		return fmt.Errorf("invalid document supplied for api: %s", name)
	}

	var routes []interface{}
	for match, pathItem := range openapiDoc.Paths {
		route, err := ociOperation(match, pathItem, n.functions)
		if err != nil {
			return err
		}

		routes = append(routes, route)
	}

	// Convert to output
	routeArray := pulumi.All(routes...).ApplyT(func(vs []interface{}) []apigateway.DeploymentSpecificationRoute {
		var results []apigateway.DeploymentSpecificationRoute
		for _, v := range vs {
			results = append(results, v.(apigateway.DeploymentSpecificationRoute))
		}

		return results
	}).(apigateway.DeploymentSpecificationRouteArrayOutput)

	gatewayName := fmt.Sprintf("gateway-%s", name)
	gateway, err := apigateway.NewGateway(ctx, gatewayName, &apigateway.GatewayArgs{
		DisplayName:   pulumi.String(gatewayName),
		CompartmentId: n.compartment.CompartmentId,
		EndpointType:  pulumi.String("PUBLIC"),
		SubnetId:      n.subnet.ID(),
	}, opts...)
	if err != nil {
		return err
	}

	n.apis[name], err = apigateway.NewDeployment(ctx, name, &apigateway.DeploymentArgs{
		CompartmentId: n.compartment.CompartmentId,
		GatewayId:     gateway.ID(),
		Specification: apigateway.DeploymentSpecificationArgs{
			Routes: routeArray,
		},
		PathPrefix:   pulumi.String("/"), // required
		FreeformTags: pulumi.ToMap(tags.TagsAsInterface(n.stackId, name, resources.API)),
	}, opts...)
	if err != nil {
		return err
	}

	return nil
}

func ociOperation(match string, pathItem *openapi3.PathItem, funcs map[string]*functions.Function) (apigateway.DeploymentSpecificationRoute, error) {
	route := apigateway.DeploymentSpecificationRoute{
		Path: match,
	}

	for method := range pathItem.Operations() {
		route.Methods = append(route.Methods, method)
	}

	for _, operation := range pathItem.Operations() {
		serviceName := ""
		if v, ok := operation.Extensions["x-nitric-target"]; ok {
			targetMap, isMap := v.(map[string]interface{})
			if isMap {
				serviceName, _ = targetMap["name"].(string)
			}
		}

		funcIdChan := make(chan string)
		service, ok := funcs[serviceName]
		if !ok {
			return apigateway.DeploymentSpecificationRoute{}, fmt.Errorf("could not find service %s", serviceName)
		}

		service.ID().ApplyT(func(id string) string {
			funcIdChan <- id
			return id
		})

		functionId := <-funcIdChan
		route.Backend = apigateway.DeploymentSpecificationRouteBackend{
			Type:       "ORACLE_FUNCTIONS_BACKEND",
			FunctionId: &functionId,
		}
	}

	return route, nil
}
