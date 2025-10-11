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

	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GetNodeInfo", Ordered, func() {
	var node *corev1.Node

	var resetNode = func() {
		node = NewFakeNode()
		node.SetAnnotations(map[string]string{
			"key": "value",
		})
		node.SetLabels(map[string]string{
			"key": "value",
		})
	}

	BeforeEach(func() {
		resetNode()
	})

	It("should parse valid selectors and straight values", func() {
		var result string
		for selector, expected := range map[string]string{
			"label=key":      "value",
			"annotation=key": "value",
			"plain-value":    "plain-value",
		} {
			result, err = provider.parseAnnotationLabelSelector(node, selector)
			ExpectWithOffset(1, result).To(Equal(expected))
		}
	})

	It("should return an error on invalid selector", func() {
		for _, selector := range []string{
			"label=absent", "annotation=absent", "label=", "annotation==",
		} {
			_, err = provider.parseAnnotationLabelSelector(node, selector)
			ExpectWithOffset(1, err).To(HaveOccurred())
		}
	})

	It("should return an error on empty labels or annotations", func() {
		node.SetAnnotations(map[string]string{})
		_, err = provider.parseAnnotationLabelSelector(node, "annotation=key")
		ExpectWithOffset(1, err).To(HaveOccurred())
		node.SetLabels(map[string]string{})
		_, err = provider.parseAnnotationLabelSelector(node, "label=key")
		ExpectWithOffset(1, err).To(HaveOccurred())
	})
})
