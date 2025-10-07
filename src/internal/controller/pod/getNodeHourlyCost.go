// Package pod provides pod controller functionality and cost calculations.
package pod

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func GetNodeHourlyCost(ctx context.Context, c client.Client, r record.EventRecorder, pod *corev1.Pod, node *corev1.Node) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)
	hourlyCost = -1

	nodeAnnotations := node.GetAnnotations()
	if nodeAnnotations == nil {
		nodeAnnotations = map[string]string{}
	}
	var nodeHourlyCostStr string
	var exists bool
	if nodeHourlyCostStr, exists = nodeAnnotations[AnnotationNodeHourlyCost]; !exists {
		// Node is not yet processes, requeueing the pod
		return hourlyCost, ErrRequestRequeue
	}

	if nodeHourlyCostStr == UnknownCost {
		msg := fmt.Sprintf("node %s has unknown hourly cost", pod.Spec.NodeName)
		r.Eventf(pod, corev1.EventTypeWarning, "NodeHourlyCostUnknown", msg)
		return
	}

	if hourlyCost, err = strconv.ParseFloat(nodeHourlyCostStr, 64); err != nil {
		log.Error(err, "failed to parse node's hourly cost", "nodeHourlyCost", nodeHourlyCostStr)
		return
	}

	return
}
