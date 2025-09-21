package manual

import (
	"context"
	"fmt"
	"strings"

	. "github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

// If annotation is a selector for node or label - unpack
func parseAnnotationLabelSelector(node *corev1.Node, value string) (result string, err error) {
	var m map[string]string
	var selector string
	var selectorType string

	if strings.HasPrefix(value, "label=") {
		m = node.GetLabels()
		selector = strings.Split(value, "label=")[1]
		selectorType = "label"
	} else if strings.HasPrefix(value, "annotation=") {
		m = node.GetAnnotations()
		selector = strings.Split(value, "annotation=")[1]
		selectorType = "annotation"
	} else {
		// If there is no selector prefix - return value as it is, since it is a literal
		return value, err
	}

	if m == nil {
		m = map[string]string{}
	}

	// Try find referenced label or annotation
	var exists bool
	if value, exists = m[selector]; !exists {
		return value, fmt.Errorf("could not find %s %s", selectorType, selector)
	}
	return
}

func GetNodeInfo(ctx context.Context, r record.EventRecorder, node *corev1.Node) (info NodeInfo, err error) {
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
	if info.Capacity, err = parseAnnotationLabelSelector(node, info.Capacity); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "CapacityGetError", err.Error())
		return
	}

	// Get node's type from labels/annotations
	if info.Type, exists = annotations[AnnotationNodeType]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoType", fmt.Sprintf("%s is not defined", AnnotationNodeType))
		return
	}
	if info.Type, err = parseAnnotationLabelSelector(node, info.Type); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "TypeGetError", err.Error())
		return
	}

	// Get node's availability zone from labels/annotations
	if info.AvailabilityZone, exists = annotations[AnnotationNodeAvailabilityZone]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoAvailabilityZone", fmt.Sprintf("%s is not defined", AnnotationNodeAvailabilityZone))
		return
	}
	if info.AvailabilityZone, err = parseAnnotationLabelSelector(node, info.AvailabilityZone); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "AvailabilityZoneGetError", err.Error())
		return
	}

	return
}
