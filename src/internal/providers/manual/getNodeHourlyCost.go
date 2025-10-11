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

// Package manual provides manual provider functionality for the controller.
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

func (*Provider) GetNodeHourlyCost(ctx context.Context, r record.EventRecorder, node *corev1.Node) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)

	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	var hourlyCostStr string
	var exists bool
	if hourlyCostStr, exists = annotations[AnnotationNodeHourlyCost]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoHourlyCost", "no provider for this .spec.providerID implemented and no manual hourly cost set")
		hourlyCost = -1
		return
	}

	if hourlyCostStr == UnknownCost {
		msg := fmt.Sprintf("node %s has unknown hourly cost", node.Name)
		r.Eventf(node, corev1.EventTypeWarning, "NodeHourlyCostUnknown", msg)
		return
	}

	if hourlyCost, err = strconv.ParseFloat(hourlyCostStr, 64); err != nil {
		msg := fmt.Sprintf("failed to parse the node price: %s", hourlyCostStr)
		log.Error(err, msg)
		return
	}

	return
}
