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

// Package pod provides Kubernetes controller implementations for cost management.
package pod

import (
	"context"
	"time"

	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/internal/utils"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	Reconciler
}

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;list;update;patch
// +kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list;watch
// +kubebuilder:rbac:groups="batch",resources=cronjobs,verbs=get;list;watch
// +kubebuilder:rbac:groups=metrics.k8s.io,resources=pods,verbs=get;list;watch

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := logf.FromContext(ctx)

	pod := corev1.Pod{}
	if err = r.Get(ctx, req.NamespacedName, &pod); err != nil {
		// Object does not exist, ignore the event and return
		if !errors.IsNotFound(err) {
			log.Error(err, "cannot get the pod")
		}
		return result, client.IgnoreNotFound(err)
	}
	log = log.WithValues("pod", pod.Name)

	// Pod is not yet scheduled
	if pod.Spec.NodeName == "" {
		return RequeueResult, err
	}

	// Handle deletion
	if pod.GetDeletionTimestamp() != nil {
		deletePodMetrics(&pod)
		return
	}

	// Get pod's node
	node := corev1.Node{}
	if err = r.Get(ctx, types.NamespacedName{Name: pod.Spec.NodeName}, &node); err != nil {
		// Object does not exist, ignore the event and return
		if !errors.IsNotFound(err) {
			log.Error(err, "cannot get the node")
		}
		return result, client.IgnoreNotFound(err)
	}

	// Get pod info
	var info PodInfo

	// Get node's hourly cost
	if info.NodeHourlyCost, err = r.getNodeHourlyCost(ctx, &node); err != nil {
		if CheckRequeue(err) {
			err = nil
			return RequeueResult, err
		}
		return
	}
	// If cost is unknown
	if info.NodeHourlyCost < 0 {
		return
	}

	// Calculate node's reference costs
	info.NodeCPUCoreHourlyCost, info.NodeMemoryMiBHourlyCost = r.getResourcesRefHourlyCost(&node, info.NodeHourlyCost)

	// Calculate minimum pod hourly cost basing on resources requests
	info.PodRequestsHourlyCost = r.getRequestsHourlyCost(ctx, &pod, info.NodeCPUCoreHourlyCost, info.NodeMemoryMiBHourlyCost)

	// Get owner
	if len(pod.GetOwnerReferences()) > 0 {
		ownerRef := pod.GetOwnerReferences()[0]
		info.Owner.Kind = ownerRef.Kind
		info.Owner.Name = ownerRef.Name
		// Get Deployment name for ReplicaSet
		if ownerRef.Kind == "ReplicaSet" {
			replicaset := appsv1.ReplicaSet{}
			if err = r.Get(ctx, types.NamespacedName{Namespace: pod.Namespace, Name: ownerRef.Name}, &replicaset); err != nil {
				// Object does not exist, ignore the event and return
				if !errors.IsNotFound(err) {
					log.Error(err, "cannot get the replicaset")
				}
				return result, client.IgnoreNotFound(err)
			}
			// Copy ReplicaSet owner to pod info
			if len(replicaset.GetOwnerReferences()) > 0 {
				ownerRef = replicaset.GetOwnerReferences()[0]
				info.Owner.Kind = ownerRef.Kind
				info.Owner.Name = ownerRef.Name
			}
		}
	}

	// Update metrics
	createPodMetrics(&pod, &info)

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Register index: spec.nodeName → pod
	if err := mgr.GetFieldIndexer().IndexField(context.Background(),
		&corev1.Pod{}, ".spec.nodeName",
		func(obj client.Object) []string {
			pod := obj.(*corev1.Pod)
			if pod.Spec.NodeName == "" {
				return nil
			}
			return []string{pod.Spec.NodeName}
		}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		// Watch Nodes in case of cost update and enqueue...
		// ... for reconciliation only required Pods
		Watches(
			&corev1.Node{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				node := obj.(*corev1.Node)
				var requests []reconcile.Request
				// Reconcile only Pods on the Nodes that had their price updated recently
				if value, exists := node.GetAnnotations()[AnnotationCostUpdatedAt]; exists {
					// Parse the time from the annotation value...
					t, err := time.Parse(time.RFC3339, value)
					if err != nil {
						// ... and reconcile nothing if failed to parse
						return requests
					}
					// Reconcile if price has been updated recently
					if time.Since(t).Seconds() < 10 {
						// List Pods on this node
						var pods corev1.PodList
						if err := r.Client.List(ctx, &pods, client.MatchingFields{".spec.nodeName": node.Name}); err != nil {
							return nil
						}
						// Prepare reconciliation requests
						for _, pod := range pods.Items {
							requests = append(requests, reconcile.Request{
								NamespacedName: types.NamespacedName{
									Namespace: pod.Namespace,
									Name:      pod.Name,
								},
							})
						}
					}
				}

				return requests
			}),
		).
		Named("pod").
		Complete(r)
}
