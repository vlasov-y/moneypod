// Copyright 2025 The MoneyPod Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package node provides Kubernetes controller implementations for cost management.
package node

import (
	"context"
	"time"

	. "github.com/vlasov-y/moneypod/internal/providers"
	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/internal/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	Reconciler
}

// +kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=nodes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;get;list;patch;update;watch

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := logf.FromContext(ctx)

	node := corev1.Node{}
	if err = r.Get(ctx, req.NamespacedName, &node); err != nil {
		// Object does not exist, ignore the event and return
		if !errors.IsNotFound(err) {
			msg := "cannot get the node"
			log.Error(err, msg)
		}
		return result, client.IgnoreNotFound(err)
	}
	log = log.WithValues("node", node.Name)

	// Handle deletion
	if node.GetDeletionTimestamp() != nil {
		deleteNodeMetrics(&node)
		return
	}

	// Skip not ready nodes
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady {
			if c.Status != corev1.ConditionTrue {
				log.V(1).Info("node is not yet ready")
				return RequeueResult, err
			}
			break
		}
	}

	// Manage hourly cost
	var hourlyCost float64
	if hourlyCost, err = r.updateHourlyCost(ctx, &node); err != nil {
		if CheckRequeue(err) {
			err = nil
			return RequeueResult, err
		}
		return
	}
	// If cost is unknown
	if hourlyCost < 0 {
		return
	}

	// First time - get full node info
	var info NodeInfo
	provider := NewProvider(&node)
	if info, err = provider.GetNodeInfo(ctx, r.Recorder, &node); err != nil {
		if CheckRequeue(err) {
			err = nil
			return RequeueResult, err
		}
		return
	}
	// And create metrics
	createNodeMetrics(&node, hourlyCost, &info)

	// Periodic cost refresh
	return ctrl.Result{RequeueAfter: time.Hour}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Named("node").
		Complete(r)
}
