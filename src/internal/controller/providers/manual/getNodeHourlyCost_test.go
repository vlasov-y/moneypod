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
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

var _ = Describe("GetNodeHourlyCost", Ordered, func() {
	var provider Provider
	var recorder *record.FakeRecorder
	var node *corev1.Node
	var ctx context.Context
	var resetNode = func() {
		node = NewFakeNode()
		node.SetAnnotations(map[string]string{
			AnnotationNodeHourlyCost: "0.3456",
		})
	}

	BeforeAll(func() {
		provider = Provider{}
		ctx = context.Background()
	})

	BeforeEach(func() {
		recorder = record.NewFakeRecorder(1)
		resetNode()
	})

	It("should parse a valid hourly cost", func() {
		var hourlyCost float64
		hourlyCost, err = provider.GetNodeHourlyCost(ctx, recorder, node)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		if len(recorder.Events) > 0 {
			ExpectWithOffset(3, recorder.Events).To(BeEmpty(),
				fmt.Sprintf("unexpected event is emitted: %s", <-recorder.Events))
		}
		Expect(hourlyCost).To(Equal(0.3456))
	})

	It("should return an event for absent hourly cost annotation", func() {
		annotations := node.GetAnnotations()
		delete(annotations, AnnotationNodeHourlyCost)
		node.SetAnnotations(annotations)
		_, err = provider.GetNodeHourlyCost(ctx, recorder, node)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		ExpectWithOffset(2, recorder.Events).To(HaveLen(1), "no event emitted")
		Expect(<-recorder.Events).To(ContainSubstring("NoHourlyCost"))
	})

	It("should return an event for unknown hourly cost", func() {
		annotations := node.GetAnnotations()
		annotations[AnnotationNodeHourlyCost] = "unknown"
		node.SetAnnotations(annotations)
		_, err = provider.GetNodeHourlyCost(ctx, recorder, node)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		ExpectWithOffset(2, recorder.Events).To(HaveLen(1), "no event emitted")
		Expect(<-recorder.Events).To(ContainSubstring("NodeHourlyCostUnknown"))
	})

	It("should return an error for invalid hourly cost", func() {
		for _, value := range []string{
			"", "custom value", " ", "1. 234", "1,0",
		} {
			annotations := node.GetAnnotations()
			annotations[AnnotationNodeHourlyCost] = value
			node.SetAnnotations(annotations)
			_, err = provider.GetNodeHourlyCost(ctx, recorder, node)
			ExpectWithOffset(1, err).To(HaveOccurred())
		}
	})
})
