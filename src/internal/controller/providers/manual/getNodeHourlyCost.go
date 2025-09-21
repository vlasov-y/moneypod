package manual

import (
	"context"
	"fmt"
	"strconv"

	. "github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func GetNodeHourlyCost(ctx context.Context, r record.EventRecorder, node *corev1.Node) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)

	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	var hourlyCostStr string
	var exists bool
	if hourlyCostStr, exists = annotations[AnnotationNodeHourlyCost]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoHourlyCost", "no provider for node provider implemented and no manual hourly cost set")
		hourlyCost = -1
		return
	}

	if hourlyCostStr == "unknown" {
		msg := fmt.Sprintf("node %s has unknown hourly cost", node.Name)
		r.Eventf(node, corev1.EventTypeWarning, "NodeHourlyCostUnknown", msg)
		return
	}

	if hourlyCost, err = strconv.ParseFloat(hourlyCostStr, 64); err != nil {
		msg := fmt.Sprintf("failed to parse the node price: %s", hourlyCostStr)
		log.V(1).Error(err, msg)
		return
	}

	return
}
