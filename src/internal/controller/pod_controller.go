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
	"fmt"
	"strconv"
	"strings"

	. "github.com/vlasov-y/moneypod/internal/types"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Config   *rest.Config
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="apps",resources=statefulsets,verbs=get;list
// +kubebuilder:rbac:groups="apps",resources=deployments,verbs=get;list
// +kubebuilder:rbac:groups="apps",resources=replicasets,verbs=get;list
// +kubebuilder:rbac:groups="batch",resources=jobs,verbs=get;list
// +kubebuilder:rbac:groups="batch",resources=cronjobs,verbs=get;list

func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	log := logf.FromContext(ctx)

	pod := corev1.Pod{}
	if err = r.Get(ctx, req.NamespacedName, &pod); err != nil {
		// Object does not exist, ignore the event and return
		if !errors.IsNotFound(err) {
			log.V(1).Error(err, "cannot get the pod")
		}
		return result, client.IgnoreNotFound(err)
	}
	log = log.WithValues("pod", pod.Name)

	// Handle deletion
	if pod.GetDeletionTimestamp() != nil {
		deletePodMetrics(&pod)
		return
	}

	// Get pod annotations
	podAnnotations := pod.GetAnnotations()
	if podAnnotations == nil {
		podAnnotations = map[string]string{}
	}

	// Pod have to have annotationNodeHourlyCost, if not - copy one from its Node
	if _, exists := podAnnotations[annotationNodeHourlyCost]; !exists {
		// Get Pod's Node
		node := corev1.Node{}
		if err = r.Get(ctx, types.NamespacedName{Name: pod.Spec.NodeName}, &node); err != nil {
			// Object does not exist, ignore the event and return
			if !errors.IsNotFound(err) {
				log.V(1).Error(err, "cannot get the node")
			}
			return result, client.IgnoreNotFound(err)
		}
		nodeAnnotations := node.GetAnnotations()
		if nodeAnnotations == nil {
			nodeAnnotations = map[string]string{}
		}
		// Node is not yet processes, requeueing the pod
		if _, exists := nodeAnnotations[annotationHourlyCost]; !exists {
			return requeue, err
		}
		// Applying new annotation
		podAnnotations[annotationNodeHourlyCost] = nodeAnnotations[annotationHourlyCost]
		pod.SetAnnotations(podAnnotations)
		if err = r.Update(ctx, &pod); err != nil {
			if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
				err = nil
				log.V(1).Info("requeue because of the update conflict")
				return requeue, err
			}
			log.V(1).Error(err, "failed to update the pod object")
			r.Recorder.Eventf(&pod, corev1.EventTypeWarning, "UpdatePodFailed", err.Error())
			return
		}
	}

	// Get precalculated cost...
	var hourlyCost float64
	if podAnnotations[annotationNodeHourlyCost] == "unknown" {
		hourlyCost = -1
	} else {
		// ...if it is defined
		if hourlyCost, err = strconv.ParseFloat(podAnnotations[annotationNodeHourlyCost], 64); err != nil {
			msg := fmt.Sprintf("failed to parse the price: %s", podAnnotations[annotationNodeHourlyCost])
			log.V(1).Error(err, msg)
			// If price is broken - delete the annotation
			newAnnotations := map[string]string{}
			for k, v := range podAnnotations {
				if k != annotationNodeHourlyCost {
					newAnnotations[k] = v
				}
			}
			pod.SetAnnotations(newAnnotations)
			// Update the object
			if err = r.Update(ctx, &pod); err != nil {
				if strings.Contains(err.Error(), "please apply your changes to the latest version and try again") {
					err = nil
					log.V(1).Info("requeue because of the update conflict")
					return requeue, err
				}
				log.V(1).Error(err, "failed to update the pod object")
				r.Recorder.Eventf(&pod, corev1.EventTypeWarning, "UpdatePodFailed", err.Error())
				return
			}
			return
		}
	}

	// Get pod info
	var info PodInfo
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
					log.V(1).Error(err, "cannot get the replicaset")
				}
				return result, client.IgnoreNotFound(err)
			}
			// Copy ReplicaSet owner to pod info
			if len(replicaset.GetOwnerReferences()) > 0 {
				ownerRef = pod.GetOwnerReferences()[0]
				info.Owner.Kind = ownerRef.Kind
				info.Owner.Name = ownerRef.Name
			}
		}
	}

	// Update metrics
	updatePodMetrics(&pod, hourlyCost, &info)
	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: maxConcurrentReconciles}).
		Named("pod").
		Complete(r)
}
