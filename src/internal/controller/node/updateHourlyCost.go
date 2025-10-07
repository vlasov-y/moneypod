// Package node provides node controller functionality and cost calculations.
package node

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/vlasov-y/moneypod/internal/controller/providers"
	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func UpdateHourlyCost(ctx context.Context, c client.Client, r record.EventRecorder, node *corev1.Node) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)

	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// Calculate Node hourly cost if annotationHourlyCost is not set or unknown
	if a, exists := annotations[AnnotationNodeHourlyCost]; !exists || a == UnknownCost {
		provider := providers.NewProvider(node)
		if hourlyCost, err = provider.GetNodeHourlyCost(ctx, r, node); err != nil {
			if CheckRequeue(err) {
				return hourlyCost, ErrRequestRequeue
			}
			return
		}

		// Add respective annotation
		if hourlyCost > 0 {
			log.V(2).Info("hourly cost is greater than zero", "hourlyCost", hourlyCost)
			annotations[AnnotationNodeHourlyCost] = strconv.FormatFloat(hourlyCost, 'f', 7, 64)
		} else {
			annotations[AnnotationNodeHourlyCost] = UnknownCost
		}
		node.SetAnnotations(annotations)
		// Update the node object
		if err = c.Update(ctx, node); err != nil {
			if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
				err = nil
				log.V(2).Info("requeue because of the update conflict")
				return hourlyCost, ErrRequestRequeue
			}
			log.V(1).Error(err, "failed to update the node object")
			r.Eventf(node, corev1.EventTypeWarning, "UpdateNodeFailed", err.Error())
			return
		}
	}
	// Get precalculated cost...
	if annotations[AnnotationNodeHourlyCost] == UnknownCost {
		hourlyCost = -1
	} else {
		// ...if it is defined
		if hourlyCost, err = strconv.ParseFloat(annotations[AnnotationNodeHourlyCost], 64); err != nil {
			msg := fmt.Sprintf("failed to parse the price: %s", annotations[AnnotationNodeHourlyCost])
			log.V(1).Error(err, msg)
			// If price is broken - delete the annotation
			newAnnotations := map[string]string{}
			for k, v := range annotations {
				if k != AnnotationNodeHourlyCost {
					newAnnotations[k] = v
				}
			}
			node.SetAnnotations(newAnnotations)
			// Update the object
			if err = c.Update(ctx, node); err != nil {
				if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
					err = nil
					log.V(2).Info("requeue because of the update conflict")
					return hourlyCost, ErrRequestRequeue
				}
				log.V(1).Error(err, "failed to update the node object")
				r.Eventf(node, corev1.EventTypeWarning, "UpdateNodeFailed", err.Error())
				return
			}
			return
		}
	}

	return
}
