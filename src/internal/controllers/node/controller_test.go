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

package node

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/vlasov-y/moneypod/internal/types"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("NodeReconciler", Ordered, func() {
	var (
		node    *corev1.Node
		nodeKey types.NamespacedName
		req     reconcile.Request
	)

	BeforeEach(func() {
		node = &corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test",
				Annotations: map[string]string{
					AnnotationNodeHourlyCost:       "10.0",
					AnnotationNodeCapacity:         "annotation=custom/capacity-type",
					AnnotationNodeType:             "label=node.kubernetes.io/instance-type",
					AnnotationNodeAvailabilityZone: "eu-central-1b",
					"custom/capacity-type":         "spot",
				},
				Labels: map[string]string{
					"node.kubernetes.io/instance-type": "t3a.2xlarge",
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
			},
		}
		nodeKey = types.NamespacedName{Name: node.Name}
		req = reconcile.Request{NamespacedName: nodeKey}

		Expect(c.Create(ctx, node)).To(Succeed())
		Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
	})

	AfterEach(func() {
		Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
		node.SetFinalizers([]string{})
		Expect(c.Update(ctx, node)).To(Succeed())
		if node.GetDeletionTimestamp() == nil {
			Expect(c.Delete(ctx, node)).To(Succeed())
		}
	})

	Context("when reconciling a ready node", func() {
		It("should handle the reconciliation successfully", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result.RequeueAfter).To(Equal(CostRefreshInterval))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
			Expect(node.Annotations).To(HaveKey(AnnotationCostUpdatedAt))
			_, err = time.Parse(time.RFC3339, node.Annotations[AnnotationCostUpdatedAt])
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
		})

		It("should handle updated-at annotation deletion", func() {
			annotations := map[string]string{}
			for k, v := range node.Annotations {
				if k != AnnotationCostUpdatedAt {
					annotations[k] = v
				}
			}
			node.SetAnnotations(annotations)
			Expect(c.Update(ctx, node)).To(Succeed())
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result.RequeueAfter).To(Equal(CostRefreshInterval))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
			Expect(node.Annotations).To(HaveKey(AnnotationCostUpdatedAt))
			_, err = time.Parse(time.RFC3339, node.Annotations[AnnotationCostUpdatedAt])
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
		})

		It("should handle broken updated-at annotation", func() {
			node.Annotations[AnnotationCostUpdatedAt] = "broken"
			Expect(c.Update(ctx, node)).To(Succeed())
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result.RequeueAfter).To(Equal(CostRefreshInterval))
			Expect(c.Get(ctx, nodeKey, node)).To(Succeed())
			Expect(node.Annotations).To(HaveKey(AnnotationCostUpdatedAt))
			_, err = time.Parse(time.RFC3339, node.Annotations[AnnotationCostUpdatedAt])
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
		})
	})

	Context("when node is not ready", func() {
		BeforeEach(func() {
			node.Status.Conditions[0].Status = corev1.ConditionFalse
			Expect(c.Status().Update(ctx, node)).To(Succeed())
		})

		It("should handle the reconciliation successfully without any events", func() {
			_, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			Expect(recorder.Events).To(BeEmpty())
		})
	})

	Context("when node does not have provider and annotations configured", func() {
		BeforeEach(func() {
			node.SetAnnotations(map[string]string{})
			Expect(c.Status().Update(ctx, node)).To(Succeed())
		})

		It("should handle the reconciliation successfully with a warning event", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
			Expect(recorder.Events).To(HaveLen(1))
			Expect(<-recorder.Events).To(ContainSubstring("Warning"))
		})
	})

	Context("when node does not exist", func() {
		It("should ignore not found errors", func() {
			nonExistentKey := types.NamespacedName{Name: "non-existent-node"}
			req := reconcile.Request{NamespacedName: nonExistentKey}
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			Expect(result).To(Equal(ctrl.Result{}))
		})
	})

	Context("when node is being deleted", func() {
		BeforeEach(func() {
			node.SetFinalizers(append(node.GetFinalizers(), "unit.test/finalizer"))
			Expect(c.Update(ctx, node)).To(Succeed())
			Expect(c.Delete(ctx, node)).To(Succeed())
		})

		It("should handle deletion gracefully", func() {
			result, err := reconciler.Reconcile(ctx, req)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())
			ExpectWithOffset(2, result).To(Equal(ctrl.Result{}))
		})
	})
})
