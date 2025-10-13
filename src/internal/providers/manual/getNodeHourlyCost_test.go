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
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/vlasov-y/moneypod/internal/types"
	. "github.com/vlasov-y/moneypod/test/utils"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GetNodeHourlyCost", Ordered, func() {
	var node *corev1.Node

	BeforeEach(func() {
		node = NewFakeNode()
		node.SetAnnotations(map[string]string{
			AnnotationNodeHourlyCost: "10.0",
		})
	})

	Context("when node has a valid cost", func() {
		It("should parse the cost successfully", func() {
			var hourlyCost float64
			hourlyCost, err = provider.GetNodeHourlyCost(ctx, recorder, node)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			if len(recorder.Events) > 0 {
				ExpectWithOffset(3, recorder.Events).To(BeEmpty(),
					fmt.Sprintf("unexpected event is emitted: %s", <-recorder.Events))
			}
			Expect(hourlyCost).To(Equal(10.0))
		})
	})

	Context("when node has no hourly cost", func() {
		BeforeEach(func() {
			delete(node.Annotations, AnnotationNodeHourlyCost)
		})

		It("should return an event and no error", func() {
			_, err = provider.GetNodeHourlyCost(ctx, recorder, node)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, recorder.Events).To(HaveLen(1), "no event emitted")
			Expect(<-recorder.Events).To(ContainSubstring("NoHourlyCost"))
		})
	})

	Context("when node has unknown hourly cost", func() {
		BeforeEach(func() {
			node.Annotations[AnnotationNodeHourlyCost] = "unknown"
		})

		It("should return an event and no error", func() {
			_, err = provider.GetNodeHourlyCost(ctx, recorder, node)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, recorder.Events).To(HaveLen(1), "no event emitted")
			Expect(<-recorder.Events).To(ContainSubstring("NodeHourlyCostUnknown"))
		})
	})

	Context("when node has broken hourly cost value", func() {
		It("should return an error", func() {
			for _, value := range []string{
				"", "custom value", " ", "1. 234", "1,0",
			} {
				By(fmt.Sprintf("Trying: %s", value))
				node.Annotations[AnnotationNodeHourlyCost] = value
				_, err = provider.GetNodeHourlyCost(ctx, recorder, node)
				ExpectWithOffset(1, err).To(HaveOccurred())
			}
		})
	})
})
