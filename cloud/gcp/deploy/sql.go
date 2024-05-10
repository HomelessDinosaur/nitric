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
	"fmt"
	"strings"

	"github.com/nitrictech/nitric/cloud/common/deploy/image"
	"github.com/nitrictech/nitric/cloud/gcp/deploy/embeds"
	deploymentspb "github.com/nitrictech/nitric/core/pkg/proto/deployments/v1"
	cloudbuild "github.com/pulumi/pulumi-google-native/sdk/go/google/cloudbuild/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

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

	databaseUrl := pulumi.Sprintf("postgres://%s:%s@%s:%s/%s", "root", a.DbMasterPassword.Result, a.MasterDb.PublicIpAddress, "5432", name)

	steps := cloudbuild.BuildStepArray{
		cloudbuild.BuildStepArgs{
			Name: pulumi.String("gcr.io/cloud-builders/git"),
			Env: pulumi.ToStringArray([]string{
				"DB_MASTER_ENDPOINT=${_DATABASE_HOST}",
				"DB_MASTER_USERNAME=${_DATABASE_USER}",
				"DB_MASTER_DATABASE_NAME=${_MASTER_DATABASE_NAME}",
				"DB_NAME=${_DATABASE_NAME}",
				"PGPASSWORD=${_DATABASE_PASSWORD}",
			}),
			Entrypoint: pulumi.String("bash"),
			Args: pulumi.ToStringArray([]string{
				"-c",
				embeds.GetCloudBuildCreateDatabaseConfig(),
			}),
		},
	}

	// If there is a migration, then add a step to run the migration image
	if config.GetImageUri() != "" && a.DatabaseMigrationBuild[name] == nil {
		steps = append(steps, cloudbuild.BuildStepArgs{
			Name: pulumi.String("gcr.io/cloud-builders/docker"),
			Env: pulumi.ToStringArray([]string{
				"NITRIC_DB_NAME=${_DATABASE_NAME}",
				"DB_URL=${_DATABASE_URL}",
			}),
			Args: pulumi.StringArray{pulumi.String("run"), image.URI()},
		})
	}

	a.DatabaseMigrationBuild[name], err = cloudbuild.NewBuild(ctx, fmt.Sprintf("%s-build", name), &cloudbuild.BuildArgs{
		Location:  pulumi.String(a.Region),
		ProjectId: pulumi.String(a.GcpConfig.ProjectId),
		Substitutions: pulumi.StringMap{
			"_DATABASE_HOST":        a.MasterDb.PublicIpAddress,
			"_DATABASE_USER":        pulumi.String("root"),
			"_MASTER_DATABASE_NAME": a.MasterDb.Name,
			"_DATABASE_NAME":        pulumi.String(name),
			"_DATABASE_PASSWORD":    a.DbMasterPassword.Result,
			"_DATABASE_URL":         databaseUrl,
		},
		Steps: steps,
	}, pulumi.DependsOn([]pulumi.Resource{a.MasterDb, image}))
	if err != nil {
		return err
	}

	return nil
}
