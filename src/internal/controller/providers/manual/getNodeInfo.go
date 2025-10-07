package manual

import (
	"context"
	"fmt"

	. "github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

func (provider *Provider) GetNodeInfo(ctx context.Context, r record.EventRecorder, node *corev1.Node) (info NodeInfo, err error) {
	var exists bool
	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// Get node's capacity from labels/annotations
	if info.Capacity, exists = annotations[AnnotationNodeCapacity]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoCapacity", fmt.Sprintf("%s is not defined", AnnotationNodeCapacity))
		return
	}
	if info.Capacity, err = provider.parseAnnotationLabelSelector(node, info.Capacity); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "CapacityGetError", err.Error())
		return
	}

	// Get node's type from labels/annotations
	if info.Type, exists = annotations[AnnotationNodeType]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoType", fmt.Sprintf("%s is not defined", AnnotationNodeType))
		return
	}
	if info.Type, err = provider.parseAnnotationLabelSelector(node, info.Type); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "TypeGetError", err.Error())
		return
	}

	// Get node's availability zone from labels/annotations
	if info.AvailabilityZone, exists = annotations[AnnotationNodeAvailabilityZone]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoAvailabilityZone", fmt.Sprintf("%s is not defined", AnnotationNodeAvailabilityZone))
		return
	}
	if info.AvailabilityZone, err = provider.parseAnnotationLabelSelector(node, info.AvailabilityZone); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "AvailabilityZoneGetError", err.Error())
		return
	}

	return
}
