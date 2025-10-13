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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *NodeReconciler) updateHourlyCost(ctx context.Context, node *corev1.Node) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)

	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	// Check hourly cost update condition transition time
	costUpdatedAt := time.Unix(0, 0)
	if value, exists := annotations[AnnotationCostUpdatedAt]; exists {
		costUpdatedAt, _ = time.Parse(time.RFC3339, value)
	}

	// Update the price only if...
	if hourlyCostStr := annotations[AnnotationNodeHourlyCost]; hourlyCostStr == UnknownCost || // 1. Last refresh was not successful
		time.Since(costUpdatedAt) > CostRefreshInterval { // 2. It is time to refresh or was never updated

		log.V(1).Info("fetching new node hourly cost")

		// Calculate Node hourly cost if annotationHourlyCost is not set or unknown
		provider := NewProvider(node)
		if hourlyCost, err = provider.GetNodeHourlyCost(ctx, r.Recorder, node); err != nil {
			if CheckRequeue(err) {
				return hourlyCost, ErrRequestRequeue
			}
			return
		}

		if hourlyCost > 0 {
			log.V(1).Info("fetched hourly cost successfully", "hourlyCost", hourlyCost)
			annotations[AnnotationNodeHourlyCost] = strconv.FormatFloat(hourlyCost, 'f', 10, 64)
			annotations[AnnotationCostUpdatedAt] = time.Now().UTC().Format(time.RFC3339)
		} else {
			log.V(1).Info("hourly cost is unknown", "hourlyCost", hourlyCost)
			annotations[AnnotationNodeHourlyCost] = UnknownCost
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
	}

	// Parse cost from annotation as float
	if annotations[AnnotationNodeHourlyCost] != UnknownCost {
		if hourlyCost, err = strconv.ParseFloat(annotations[AnnotationNodeHourlyCost], 64); err != nil {
			msg := fmt.Sprintf("failed to parse the cost: %s", annotations[AnnotationNodeHourlyCost])
			log.Error(err, msg)
			// If price is broken - set cost to unknown
			annotations[AnnotationNodeHourlyCost] = UnknownCost
			node.SetAnnotations(annotations)
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
		}
	}

	// Handle unknown cost in one place
	if annotations[AnnotationNodeHourlyCost] == UnknownCost {
		hourlyCost = -1
	}

	return
}
