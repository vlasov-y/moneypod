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

package manual

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/vlasov-y/moneypod/test/utils"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GetNodeInfo", Ordered, func() {
	var node *corev1.Node

	BeforeEach(func() {
		node = NewFakeNode()
		node.SetAnnotations(map[string]string{
			"key": "value",
		})
		node.SetLabels(map[string]string{
			"key": "value",
		})
	})

	Context("when selector is valid", func() {
		It("should get the referenced value", func() {
			var result string
			for selector, expected := range map[string]string{
				"label=key":      "value",
				"annotation=key": "value",
				"plain-value":    "plain-value",
			} {
				By(selector)
				result, err = provider.parseAnnotationLabelSelector(node, selector)
				ExpectWithOffset(1, result).To(Equal(expected))
			}
		})
	})

	Context("when selector is broken", func() {
		It("should return an error", func() {
			for _, selector := range []string{
				"label=absent", "annotation=absent", "label=", "annotation==",
			} {
				By(selector)
				_, err = provider.parseAnnotationLabelSelector(node, selector)
				ExpectWithOffset(1, err).To(HaveOccurred())
			}
		})
	})

	Context("when there is no annotations or labels", func() {
		It("should return an error", func() {
			node.SetAnnotations(map[string]string{})
			_, err = provider.parseAnnotationLabelSelector(node, "annotation=key")
			ExpectWithOffset(1, err).To(HaveOccurred())
			node.SetLabels(map[string]string{})
			_, err = provider.parseAnnotationLabelSelector(node, "label=key")
			ExpectWithOffset(1, err).To(HaveOccurred())
		})
	})

})
