// Copyright 2025 The MoneyPod Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package node provides node controller functionality and cost calculations.
package node

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	. "github.com/vlasov-y/moneypod/internal/providers"
	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *NodeReconciler) updateHourlyCost(ctx context.Context, node *corev1.Node) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)

	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// Check hourly cost update condition transition time
	var condition *corev1.NodeCondition
	for _, c := range node.Status.Conditions {
		if c.Type == ConditionNodeHourlyCost.Type {
			condition = c.DeepCopy()
		}
	}
	// Update the price only if...
	if _, annotationExists := annotations[AnnotationNodeHourlyCost]; condition == nil || // 1. Price was never updated
		time.Now().After(condition.LastHeartbeatTime.Add(time.Hour)) || // 2. It is time to refresh
		condition.Status != corev1.ConditionTrue || // 3. Last refresh was not successful
		!annotationExists { // 4. No price annotation

		log.V(1).Info("fetching new node hourly cost")

		// Calculate Node hourly cost if annotationHourlyCost is not set or unknown
		provider := NewProvider(node)
		if hourlyCost, err = provider.GetNodeHourlyCost(ctx, r.Recorder, node); err != nil {
			if CheckRequeue(err) {
				return hourlyCost, ErrRequestRequeue
			}

			// Update condition to set HourlyCost condition state to False
			err = r.updateCondition(ctx, node, corev1.NodeCondition{
				Type:               ConditionNodeHourlyCost.Type,
				Status:             corev1.ConditionFalse,
				Reason:             ConditionNodeHourlyCost.ReasonUnknown,
				Message:            err.Error(),
				LastTransitionTime: metav1.Now(),
			})
			if CheckRequeue(err) {
				return hourlyCost, ErrRequestRequeue
			}
			return
		}

		// Add respective annotation and update status condition
		if condition == nil {
			condition = &corev1.NodeCondition{
				Type:   ConditionNodeHourlyCost.Type,
				Status: corev1.ConditionUnknown,
			}
		}
		condition.LastHeartbeatTime = metav1.Now()
		if hourlyCost > 0 {
			log.V(1).Info("fetched hourly cost successfully", "hourlyCost", hourlyCost)
			annotations[AnnotationNodeHourlyCost] = strconv.FormatFloat(hourlyCost, 'f', 10, 64)
			if condition.Status != corev1.ConditionTrue {
				condition.Status = corev1.ConditionTrue
				condition.LastTransitionTime = metav1.Now()
			}
			condition.Reason = ConditionNodeHourlyCost.ReasonUpdated
			condition.Message = fmt.Sprintf("successfully fetched new hourly cost: %s", annotations[AnnotationNodeHourlyCost])
			condition.LastHeartbeatTime = metav1.Now()
		} else {
			log.V(1).Info("hourly cost is unknown", "hourlyCost", hourlyCost)
			annotations[AnnotationNodeHourlyCost] = UnknownCost
			if condition.Status != corev1.ConditionUnknown {
				condition.Status = corev1.ConditionUnknown
				condition.LastTransitionTime = metav1.Now()
			}
			condition.Reason = ConditionNodeHourlyCost.ReasonUnknown
			condition.Message = "failed to fetch hourly cost using known providers"
		}

		node.SetAnnotations(annotations)
		// Update the node object
		if err = r.Update(ctx, node); err != nil {
			if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
				err = nil
				log.V(1).Info("requeue because of the update conflict")
				return hourlyCost, ErrRequestRequeue
			}
			log.Error(err, "failed to update the node object")
			r.Recorder.Eventf(node, corev1.EventTypeWarning, "UpdateNodeFailed", err.Error())
			return
		}

		// Update condition to reschedule next price update
		if err = r.updateCondition(ctx, node, *condition); err != nil {
			if CheckRequeue(err) {
				return hourlyCost, ErrRequestRequeue
			}
		}
	}

	// Get the const from annotations...
	if annotations[AnnotationNodeHourlyCost] == UnknownCost {
		hourlyCost = -1
	} else {
		// ...if it is defined
		if hourlyCost, err = strconv.ParseFloat(annotations[AnnotationNodeHourlyCost], 64); err != nil || hourlyCost == 0 {
			if hourlyCost == 0 {
				log.Info("node hourly cost has been set to 0 so reevaluation is required")
			} else {
				msg := fmt.Sprintf("failed to parse the cost: %s", annotations[AnnotationNodeHourlyCost])
				log.Error(err, msg)
			}
			// If price is broken - delete the annotation
			newAnnotations := map[string]string{}
			for k, v := range annotations {
				if k != AnnotationNodeHourlyCost {
					newAnnotations[k] = v
				}
			}
			node.SetAnnotations(newAnnotations)
			// Update the object
			if err = r.Update(ctx, node); err != nil {
				if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
					err = nil
					log.V(1).Info("requeue because of the update conflict")
					return hourlyCost, ErrRequestRequeue
				}
				log.Error(err, "failed to update the node object")
				r.Recorder.Eventf(node, corev1.EventTypeWarning, "UpdateNodeFailed", err.Error())
				return
			}
			return
		}
	}

	return
}
