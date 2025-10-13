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

package pod

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/internal/utils"
	. "github.com/vlasov-y/moneypod/test/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("PodReconciler", Ordered, func() {
	var (
		node    *corev1.Node
		pod     *corev1.Pod
		podKey  types.NamespacedName
		nodeKey types.NamespacedName
		req     reconcile.Request
	)

	BeforeEach(func() {
		node = NewFakeNode()
		node.SetAnnotations(map[string]string{
			AnnotationNodeHourlyCost: "10.0",
		})
		nodeKey = types.NamespacedName{Name: node.Name}
		pod = NewFakePod()
		pod.Spec.NodeName = node.Name
		podKey = types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}
		req = reconcile.Request{NamespacedName: podKey}

		Expect(c.Create(ctx, node)).To(Succeed())
		Expect(c.Status().Update(ctx, node)).To(Succeed())
		Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
		node.TypeMeta = NewTypeMeta(node, c.Scheme())

		Expect(c.Create(ctx, pod)).To(Succeed())
		Expect(c.Status().Update(ctx, pod)).To(Succeed())
		Expect(c.Get(ctx, podKey, pod)).To(Succeed())
		pod.TypeMeta = NewTypeMeta(pod, c.Scheme())
	})

	AfterEach(func() {
		if pod != nil {
			Expect(c.Get(ctx, podKey, pod)).To(Succeed())
			pod.SetFinalizers([]string{})
			Expect(c.Update(ctx, pod)).To(Succeed())
			if pod.GetDeletionTimestamp() == nil {
				Expect(c.Delete(ctx, pod)).To(Succeed())
			}
		}
		if node != nil {
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
			node.SetFinalizers([]string{})
			Expect(c.Update(ctx, node)).To(Succeed())
			if node.GetDeletionTimestamp() == nil {
				Expect(c.Delete(ctx, node)).To(Succeed())
			}
		}
	})

	Context("when reconciling a scheduled pod", func() {
		It("should handle the reconciliation successfully", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
		})

		It("should handle pod with owner reference", func() {
			By("creating a statefulset")
			statefulset := NewFakeStatefulSet()
			statefulsetKey := types.NamespacedName{Name: statefulset.Name, Namespace: statefulset.Namespace}
			Expect(c.Create(ctx, statefulset)).To(Succeed())
			Expect(c.Status().Update(ctx, statefulset)).To(Succeed())
			Expect(c.Get(ctx, statefulsetKey, statefulset)).To(Succeed())
			statefulset.TypeMeta = NewTypeMeta(statefulset, c.Scheme())

			By("setting pod owner reference to statefulset")
			pod.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: statefulset.APIVersion,
					Kind:       statefulset.Kind,
					Name:       statefulset.Name,
					UID:        statefulset.UID,
					Controller: ptr.To(true),
				},
			})
			Expect(c.Update(ctx, pod)).To(Succeed())

			By("reconciling")
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))

			By("deleting statefulset")
			Expect(c.Delete(ctx, statefulset)).To(Succeed())
		})

		It("should handle pod with replicaset owner", func() {
			By("creating a deployment")
			deployment := NewFakeDeployment()
			deploymentKey := types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}
			Expect(c.Create(ctx, deployment)).To(Succeed())
			Expect(c.Status().Update(ctx, deployment)).To(Succeed())
			Expect(c.Get(ctx, deploymentKey, deployment)).To(Succeed())
			deployment.TypeMeta = NewTypeMeta(deployment, c.Scheme())

			By("creating a repicaset")
			replicaset := NewFakeReplicaSet()
			replicasetKey := types.NamespacedName{Name: replicaset.Name, Namespace: replicaset.Namespace}
			Expect(c.Create(ctx, replicaset)).To(Succeed())
			Expect(c.Status().Update(ctx, replicaset)).To(Succeed())
			Expect(c.Get(ctx, replicasetKey, replicaset)).To(Succeed())
			replicaset.TypeMeta = NewTypeMeta(replicaset, c.Scheme())

			By("setting replicaset owner reference to deployment")
			replicaset.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: deployment.APIVersion,
					Kind:       deployment.Kind,
					Name:       deployment.Name,
					UID:        deployment.UID,
					Controller: ptr.To(true),
				},
			})
			Expect(c.Update(ctx, replicaset)).To(Succeed())

			By("setting pod owner reference to replicaset")
			pod.SetOwnerReferences([]metav1.OwnerReference{
				{
					APIVersion: replicaset.APIVersion,
					Kind:       replicaset.Kind,
					Name:       replicaset.Name,
					UID:        replicaset.UID,
					Controller: ptr.To(true),
				},
			})
			Expect(c.Update(ctx, pod)).To(Succeed())

			By("reconciling")
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))

			By("deleting deployment and replicaset")
			Expect(c.Delete(ctx, replicaset)).To(Succeed())
			Expect(c.Delete(ctx, deployment)).To(Succeed())
		})
	})

	Context("when pod is not scheduled", func() {
		BeforeEach(func() {
			Expect(c.Delete(ctx, pod)).To(Succeed())
			pod.ResourceVersion = ""
			pod.Spec.NodeName = ""
			Expect(c.Create(ctx, pod)).To(Succeed())
		})

		It("should requeue the reconciliation", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(RequeueResult))
		})
	})

	Context("when node has unknown cost", func() {
		BeforeEach(func() {
			node.Annotations[AnnotationNodeHourlyCost] = UnknownCost
			Expect(c.Update(ctx, node)).To(Succeed())
		})

		It("should handle the reconciliation without error", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
		})
	})

	Context("when pod does not exist", func() {
		It("should ignore not found errors", func() {
			nonExistentKey := types.NamespacedName{Name: "non-existent-pod", Namespace: "default"}
			req := reconcile.Request{NamespacedName: nonExistentKey}
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Context("when node does not exist", func() {
		BeforeEach(func() {
			Expect(c.Delete(ctx, pod)).To(Succeed())
			pod.ResourceVersion = ""
			pod.Spec.NodeName = "non-existent-node"
			Expect(c.Create(ctx, pod)).To(Succeed())
		})

		It("should ignore not found errors", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Context("when pod is being deleted", func() {
		BeforeEach(func() {
			pod.SetFinalizers(append(pod.GetFinalizers(), "unit.test/finalizer"))
			Expect(c.Update(ctx, pod)).To(Succeed())
			Expect(c.Delete(ctx, pod)).To(Succeed())
		})

		It("should handle deletion gracefully", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
		})
	})
})
