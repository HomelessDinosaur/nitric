// Copyright Nitric Pty Ltd.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deploy

import (
	"context"
	"fmt"
	"strings"
	"time"

	cloudbuild "cloud.google.com/go/cloudbuild/apiv1/v2"
	"cloud.google.com/go/cloudbuild/apiv1/v2/cloudbuildpb"
	"github.com/avast/retry-go"
	"github.com/nitrictech/nitric/cloud/common/deploy/image"
	deploymentspb "github.com/nitrictech/nitric/core/pkg/proto/deployments/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/sql"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func checkBuildStatus(ctx context.Context, build *cloudbuild.CreateBuildOperation) func() error {
	return func() error {
		_, err := build.Poll(ctx)
		return err
	}
}

func (a *NitricGcpPulumiProvider) SqlDatabase(ctx *pulumi.Context, parent pulumi.Resource, name string, config *deploymentspb.SqlDatabase) error {
	// Get the image name:tag from the uri
	imageUriSplit := strings.Split(config.GetImageUri(), "/")
	imageName := imageUriSplit[len(imageUriSplit)-1]

	image, err := image.NewLocalImage(ctx, name, &image.LocalImageArgs{
		RepositoryUrl: pulumi.Sprintf("gcr.io/%s/%s", a.GcpConfig.ProjectId, imageName),
		SourceImage:   config.GetImageUri(),
		Username:      pulumi.String("oauth2accesstoken"),
		Password:      pulumi.String(a.AuthToken.AccessToken),
		Server:        pulumi.String("https://gcr.io"),
	})
	if err != nil {
		return err
	}

	_, err = sql.NewDatabase(ctx, name, &sql.DatabaseArgs{
		Name:           pulumi.String(name),
		Instance:       a.masterDb.Name,
		DeletionPolicy: pulumi.String("DELETE"),
		Project:        pulumi.String(a.GcpConfig.ProjectId),
	}, pulumi.Parent(parent), pulumi.DependsOn([]pulumi.Resource{a.masterDb}))
	if err != nil {
		return err
	}

	creds, err := google.FindDefaultCredentials(ctx.Context())
	if err != nil {
		return err
	}

	client, err := cloudbuild.NewClient(ctx.Context(), option.WithCredentials(creds))
	if err != nil {
		return err
	}

	defer client.Close()

	databaseUrl := pulumi.Sprintf("postgres://%s:%s@%s:%s/%s", "postgres", a.dbMasterPassword.Result, a.masterDb.PrivateIpAddress, "5432", name)

	pulumi.All(databaseUrl, a.cloudBuildWorkerPool.ID().ToStringOutput(), image.URI()).ApplyT(func(args []interface{}) (bool, error) {
		url := args[0].(string)
		workerPoolId := args[1].(string)
		imageUri := args[2].(string)

		build, err := client.CreateBuild(ctx.Context(), &cloudbuildpb.CreateBuildRequest{
			Parent:    fmt.Sprintf("projects/%s/locations/%s", a.GcpConfig.ProjectId, a.Region),
			ProjectId: a.GcpConfig.ProjectId,
			Build: &cloudbuildpb.Build{
				Substitutions: map[string]string{
					"_DATABASE_NAME": name,
					"_DATABASE_URL":  url,
				},
				Steps: []*cloudbuildpb.BuildStep{
					{
						Name: imageUri,
						Dir:  "/",
						Env: []string{
							"NITRIC_DB_NAME=${_DATABASE_NAME}",
							"DB_URL=${_DATABASE_URL}",
						},
					},
				},
				Options: &cloudbuildpb.BuildOptions{
					Pool: &cloudbuildpb.BuildOptions_PoolOption{
						Name: workerPoolId,
					},
				},
			},
		})
		if err != nil {
			return false, err
		}

		err = retry.Do(checkBuildStatus(ctx.Context(), build), retry.Attempts(10), retry.Delay(time.Second*15))
		if err != nil {
			return false, err
		}

		a.DatabaseMigrationBuild[name] = build

		return true, nil
	})

	return nil
}
