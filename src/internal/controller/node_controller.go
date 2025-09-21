/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"strings"

	. "github.com/vlasov-y/moneypod/internal/controller/node"
	"github.com/vlasov-y/moneypod/internal/controller/providers/aws"
	"github.com/vlasov-y/moneypod/internal/controller/providers/manual"
	. "github.com/vlasov-y/moneypod/internal/types"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Config   *rest.Config
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
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
			log.V(1).Error(err, msg)
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
				log.V(2).Info("node is not yet ready")
				return requeue, err
			}
			break
		}
	}

	// Manage hourly cost
	var hourlyCost float64
	if hourlyCost, err = UpdateHourlyCost(ctx, r.Client, r.Recorder, &node); err != nil {
		if err.Error() == "requeue" {
			err = nil
			return requeue, err
		}
		return
	}
	// If cost is unknown
	if hourlyCost < 0 {
		return
	}

	// First time - get full node info
	var info NodeInfo
	if strings.HasPrefix(node.Spec.ProviderID, "aws://") {
		if info, err = aws.GetNodeInfo(ctx, r.Recorder, &node); err != nil {
			if err.Error() == "requeue" {
				err = nil
				return requeue, err
			}
			return
		}
	} else {
		if info, err = manual.GetNodeInfo(ctx, r.Recorder, &node); err != nil {
			if err.Error() == "requeue" {
				err = nil
				return requeue, err
			}
			return
		}
	}
	// And create metrics
	createNodeMetrics(&node, hourlyCost, &info)

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: MaxConcurrentReconciles}).
		Named("node").
		Complete(r)
}
