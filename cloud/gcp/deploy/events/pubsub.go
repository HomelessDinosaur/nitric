package events

import (
	common "github.com/nitrictech/nitric/cloud/common/deploy/tags"
	"github.com/nitrictech/nitric/cloud/gcp/deploy/exec"
	v1 "github.com/nitrictech/nitric/core/pkg/api/nitric/deploy/v1"
	"github.com/pkg/errors"
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

	Topic *v1.Topic
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

	return res, nil
}

type PubSubSubscription struct {
	pulumi.ResourceState

	Name string

	Subscription *pubsub.Subscription
}

type PubSubSubscriptionArgs struct {
	Function *exec.CloudRunner
	Topic string
	Url pulumi.StringInput
	ServiceAccount pulumi.StringOutput
}

func NewPubSubSubscription(ctx *pulumi.Context, name string, args *PubSubSubscriptionArgs, opts ...pulumi.ResourceOption) (*PubSubSubscription, error) {
	res := &PubSubSubscription{
		Name: name,
	}

	s, err := pubsub.NewSubscription(ctx, name, &pubsub.SubscriptionArgs{
		Topic:              pulumi.String(args.Topic),
		AckDeadlineSeconds: pulumi.Int(300),
		RetryPolicy: pubsub.SubscriptionRetryPolicyArgs{
			MinimumBackoff: pulumi.String("15s"),
			MaximumBackoff: pulumi.String("600s"),
		},
		PushConfig: pubsub.SubscriptionPushConfigArgs{
			OidcToken: pubsub.SubscriptionPushConfigOidcTokenArgs{
				ServiceAccountEmail: args.ServiceAccount,
			},
			PushEndpoint: args.Url,
		},
	}, append(opts, pulumi.Parent(args.Function))...)
	if err != nil {
		return nil, errors.WithMessage(err, "subscription "+name+"-sub")
	}

	res.Subscription = s

	return res, nil
}