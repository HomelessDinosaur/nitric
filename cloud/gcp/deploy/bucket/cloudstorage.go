package bucket

import (
	common "github.com/nitrictech/nitric/cloud/common/deploy/tags"
	v1 "github.com/nitrictech/nitric/core/pkg/api/nitric/deploy/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/storage"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CloudStorageBucket struct {
	pulumi.ResourceState

	Name string
	CloudStorage *storage.Bucket
}

type CloudStorageBucketArgs struct {
	Location string
	StackID pulumi.StringInput
	ProjectId string

	Bucket *v1.Bucket
}

func NewCloudStorageBucket(ctx *pulumi.Context, name string, args *CloudStorageBucketArgs, opts ...pulumi.ResourceOption) (*CloudStorageBucket, error) {
	res := &CloudStorageBucket {
		Name: name,
	}

	err := ctx.RegisterComponentResource("nitric:bucket:GCPCloudStorage", name, res, opts...)
	if err != nil {
		return nil, err
	}


	res.CloudStorage, err = storage.NewBucket(ctx, name, &storage.BucketArgs{
		Location: pulumi.String(args.Location),
		Project: pulumi.String(args.ProjectId),
		Labels: common.Tags(ctx, args.StackID, name),
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}