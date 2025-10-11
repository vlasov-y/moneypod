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

package node

import (
	"context"
	"fmt"
	"strings"

	. "github.com/vlasov-y/moneypod/internal/utils"

	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// updateCondition updates or replaces a condition and updates the status
func (r *NodeReconciler) updateCondition(ctx context.Context,
	node *corev1.Node, condition corev1.NodeCondition) (err error) {
	log := logf.FromContext(ctx)

	condition.Message = fmt.Sprintf("moneypod %s", condition.Message)
	found := false
	for i, c := range node.Status.Conditions {
		if c.Type == condition.Type {
			node.Status.Conditions[i] = condition
			found = true
		}
	}
	if !found {
		node.Status.Conditions = append(node.Status.Conditions, condition)
	}

	// Update the node object
	if err = r.Status().Update(ctx, node); err != nil {
		if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
			err = nil
			log.V(1).Info("requeue because of the update conflict")
			return ErrRequestRequeue
		}
		log.Error(err, "failed to update node's status object")
		r.Recorder.Eventf(node, corev1.EventTypeWarning, "UpdateStatusFailed", err.Error())
	}

	return
}
