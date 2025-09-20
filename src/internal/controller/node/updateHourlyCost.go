package pod

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/vlasov-y/moneypod/internal/controller/providers/aws"
	. "github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func UpdateHourlyCost(ctx context.Context, c client.Client, r record.EventRecorder, node *corev1.Node) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)
	hourlyCost = -1

	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// Calculate Node hourly cost if annotationHourlyCost is not set or unknown
	if a, exists := annotations[AnnotationHourlyCost]; !exists || a == "unknown" {
		// Switch by provider
		if strings.HasPrefix(node.Spec.ProviderID, "aws://") {
			if hourlyCost, err = aws.GetNodeHourlyCost(ctx, r, node); err != nil {
				if err.Error() == "requeue" {
					err = nil
					return hourlyCost, errors.New("requeue")
				}
				return
			}
		} else {
			// If no provider is implemented - set cost to unknown
			log.V(1).Info("no cost provider implemented", "providerId", node.Spec.ProviderID)
			hourlyCost = -1
		}

		// Add respective annotation
		if hourlyCost > 0 {
			annotations[AnnotationHourlyCost] = strconv.FormatFloat(hourlyCost, 'f', 7, 64)
		} else {
			annotations[AnnotationHourlyCost] = "unknown"
		}
		node.SetAnnotations(annotations)
		// Update the node object
		if err = c.Update(ctx, node); err != nil {
			if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
				err = nil
				log.V(1).Info("requeue because of the update conflict")
				return hourlyCost, errors.New("requeue")
			}
			log.V(1).Error(err, "failed to update the node object")
			r.Eventf(node, corev1.EventTypeWarning, "UpdateNodeFailed", err.Error())
			return
		}
	}
	// Get precalculated cost...
	if annotations[AnnotationHourlyCost] == "unknown" {
		hourlyCost = -1
	} else {
		// ...if it is defined
		if hourlyCost, err = strconv.ParseFloat(annotations[AnnotationHourlyCost], 64); err != nil {
			msg := fmt.Sprintf("failed to parse the price: %s", annotations[AnnotationHourlyCost])
			log.V(1).Error(err, msg)
			// If price is broken - delete the annotation
			newAnnotations := map[string]string{}
			for k, v := range annotations {
				if k != AnnotationHourlyCost {
					newAnnotations[k] = v
				}
			}
			node.SetAnnotations(newAnnotations)
			// Update the object
			if err = c.Update(ctx, node); err != nil {
				if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
					err = nil
					log.V(1).Info("requeue because of the update conflict")
					return hourlyCost, errors.New("requeue")
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
