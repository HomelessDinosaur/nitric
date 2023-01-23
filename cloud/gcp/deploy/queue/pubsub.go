package queue

import (
	common "github.com/nitrictech/nitric/cloud/common/deploy/tags"
	v1 "github.com/nitrictech/nitric/core/pkg/api/nitric/deploy/v1"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/pubsub"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type PubSubTopic struct {
	pulumi.ResourceState

	Name string
	PubSub *pubsub.Topic
}

type PubSubTopicArgs struct {
	Location string
	StackID pulumi.StringInput
	ProjectId string

	Queue *v1.Queue
}

func NewPubSubTopic(ctx *pulumi.Context, name string, args *PubSubTopicArgs, opts ...pulumi.ResourceOption) (*PubSubTopic, error) {
	res := &PubSubTopic {
		Name: name,
	}

	err := ctx.RegisterComponentResource("nitric:topic:GCPPubSubTopic", name, res, opts...)
	if err != nil {
		return nil, err
	}

	res.PubSub, err = pubsub.NewTopic(ctx, name, &pubsub.TopicArgs{
		Name: pulumi.String(name),
		Labels: common.Tags(ctx, args.StackID, name),
	})
	if err != nil {
		return nil, err
	}

	_, err = pubsub.NewSubscription(ctx, name+"-sub", &pubsub.SubscriptionArgs{
		Name:   pulumi.Sprintf("%s-nitricqueue", name),
		Topic:  res.PubSub.Name,
		Labels: common.Tags(ctx, args.StackID, name+"-sub"),
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}