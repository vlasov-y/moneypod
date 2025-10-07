package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	ec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (provider *Provider) GetNodeInfo(ctx context.Context, r record.EventRecorder, node *corev1.Node) (info types.NodeInfo, err error) {
	log := logf.FromContext(ctx)

	// Gather node information
	var instanceID string
	if instanceID, err = provider.getInstanceID(ctx, r, node); err != nil {
		return
	}
	info.ID = instanceID

	// Authorize AWS
	var awsConfig aws.Config
	if awsConfig, err = config.LoadDefaultConfig(ctx); err != nil {
		log.V(1).Error(err, "failed to load AWS config")
		return
	}
	clientEc2 := ec2.NewFromConfig(awsConfig)

	// Describe the instance
	var describe *ec2.DescribeInstancesOutput
	if describe, err = clientEc2.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		InstanceIds: []string{instanceID},
	}); err != nil {
		log.V(1).Error(err, "failed to describe the instance")
		r.Eventf(node, corev1.EventTypeWarning, "DescribeEC2InstanceFailed", err.Error())
		return
	}

	// Get remaining info
	for _, reservation := range describe.Reservations {
		for _, instance := range reservation.Instances {
			info.Capacity = string(types.OnDemand)
			info.Type = string(instance.InstanceType)
			info.AvailabilityZone = *instance.Placement.AvailabilityZone
			if instance.SpotInstanceRequestId != nil {
				info.Capacity = string(types.Spot)
			}
		}
	}

	return
}
