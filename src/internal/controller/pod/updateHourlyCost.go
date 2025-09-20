package pod

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	. "github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func UpdateHourlyCost(ctx context.Context, c client.Client, r record.EventRecorder, pod *corev1.Pod) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)
	hourlyCost = -1

	annotations := pod.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// We have to copy node cost to the pod if it does not have one yet...
	// ...or cost is unknown (hoping it will be defined)
	if a, exists := annotations[AnnotationNodeHourlyCost]; !exists || a == "unknown" {
		// Get Pod's Node
		node := corev1.Node{}
		if err = c.Get(ctx, types.NamespacedName{Name: pod.Spec.NodeName}, &node); err != nil {
			// Object does not exist, ignore the event and return
			if !apierrors.IsNotFound(err) {
				log.V(1).Error(err, "cannot get the node")
			}
			return hourlyCost, client.IgnoreNotFound(err)
		}
		nodeAnnotations := node.GetAnnotations()
		if nodeAnnotations == nil {
			nodeAnnotations = map[string]string{}
		}
		// Node is not yet processes, requeueing the pod
		if _, exists := nodeAnnotations[AnnotationHourlyCost]; !exists {
			return hourlyCost, errors.New("requeue")
		}
		// Applying new annotation
		annotations[AnnotationNodeHourlyCost] = nodeAnnotations[AnnotationHourlyCost]
		pod.SetAnnotations(annotations)
		if err = c.Update(ctx, pod); err != nil {
			if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
				err = nil
				log.V(1).Info("requeue because of the update conflict")
				return hourlyCost, errors.New("requeue")
			}
			log.V(1).Error(err, "failed to update the pod object")
			r.Eventf(pod, corev1.EventTypeWarning, "UpdatePodFailed", err.Error())
			return
		}
	}

	// Now parse existing annotation...
	if annotations[AnnotationNodeHourlyCost] == "unknown" {
		hourlyCost = -1
	} else {
		// ...if it is defined
		if hourlyCost, err = strconv.ParseFloat(annotations[AnnotationNodeHourlyCost], 64); err != nil {
			msg := fmt.Sprintf("failed to parse the price: %s", annotations[AnnotationNodeHourlyCost])
			log.V(1).Error(err, msg)
			// If value is broken - delete the annotation
			newAnnotations := map[string]string{}
			for k, v := range annotations {
				if k != AnnotationNodeHourlyCost {
					newAnnotations[k] = v
				}
			}
			pod.SetAnnotations(newAnnotations)
			// Update the object
			if err = c.Update(ctx, pod); err != nil {
				if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
					err = nil
					log.V(1).Info("requeue because of the update conflict")
					return hourlyCost, errors.New("requeue")
				}
				log.V(1).Error(err, "failed to update the pod object")
				r.Eventf(pod, corev1.EventTypeWarning, "UpdatePodFailed", err.Error())
				return
			}
			return
		}
	}

	return
}
