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

// Package pod provides pod controller functionality and cost calculations.
package pod

import (
	"context"
	"strconv"

	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/internal/utils"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *PodReconciler) getNodeHourlyCost(ctx context.Context, node *corev1.Node) (hourlyCost float64, err error) {
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
		return
	}

	if hourlyCost, err = strconv.ParseFloat(nodeHourlyCostStr, 64); err != nil {
		log.Error(err, "failed to parse node's hourly cost", "nodeHourlyCost", nodeHourlyCostStr)
		return
	}

	return
}
