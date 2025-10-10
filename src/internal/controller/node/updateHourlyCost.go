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
		// Rounding cost to 10 chars before comparing with 0
		hourlyCost, _ = strconv.ParseFloat(strconv.FormatFloat(hourlyCost, 'f', 10, 64), 64)
		// Add respective annotation
		if hourlyCost > 0 {
			log.V(1).Info("hourly cost is greater than zero", "hourlyCost", hourlyCost)
			annotations[AnnotationNodeHourlyCost] = strconv.FormatFloat(hourlyCost, 'f', 10, 64)
		} else {
			annotations[AnnotationNodeHourlyCost] = UnknownCost
		}
		node.SetAnnotations(annotations)
		// Update the node object
		if err = c.Update(ctx, node); err != nil {
			if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
				err = nil
				log.V(1).Info("requeue because of the update conflict")
				return hourlyCost, ErrRequestRequeue
			}
			log.Error(err, "failed to update the node object")
			r.Eventf(node, corev1.EventTypeWarning, "UpdateNodeFailed", err.Error())
			return
		}
	}
	// Get precalculated cost...
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
			if err = c.Update(ctx, node); err != nil {
				if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
					err = nil
					log.V(1).Info("requeue because of the update conflict")
					return hourlyCost, ErrRequestRequeue
				}
				log.Error(err, "failed to update the node object")
				r.Eventf(node, corev1.EventTypeWarning, "UpdateNodeFailed", err.Error())
				return
			}
			return
		}
	}

	return
}
