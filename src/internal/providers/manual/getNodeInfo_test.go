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
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("GetNodeInfo", Ordered, func() {
	var node *corev1.Node

	var resetNode = func() {
		node = NewFakeNode()
		node.SetAnnotations(map[string]string{
			AnnotationNodeCapacity:         "annotation=custom/capacity-type",
			AnnotationNodeType:             "label=node.kubernetes.io/instance-type",
			AnnotationNodeAvailabilityZone: "eu-central-1b",
			"custom/capacity-type":         "spot",
		})
		node.SetLabels(map[string]string{
			"node.kubernetes.io/instance-type": "t3a.2xlarge",
		})
	}

	BeforeEach(func() {
		resetNode()
	})

	It("should parse all annotations without an error", func() {
		var info NodeInfo
		info, err = provider.GetNodeInfo(ctx, recorder, node)
		ExpectWithOffset(1, err).ToNot(HaveOccurred())
		if len(recorder.Events) > 0 {
			ExpectWithOffset(3, recorder.Events).To(BeEmpty(),
				fmt.Sprintf("unexpected event is emitted: %s", <-recorder.Events))
		}
		Expect(info.ID).To(Equal("manual"))
		Expect(info.AvailabilityZone).To(Equal("eu-central-1b"))
		Expect(info.Capacity).To(Equal("spot"))
		Expect(info.Type).To(Equal("t3a.2xlarge"))
	})

	It("should return an event for absent mandatory annotation", func() {
		for annotation, str := range map[string]string{
			AnnotationNodeAvailabilityZone: "NoAvailabilityZone",
			AnnotationNodeCapacity:         "NoCapacity",
			AnnotationNodeType:             "NoType",
		} {
			By(annotation)
			annotations := node.GetAnnotations()
			delete(annotations, annotation)
			node.SetAnnotations(annotations)
			_, err = provider.GetNodeInfo(ctx, recorder, node)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, recorder.Events).To(HaveLen(1), "no event emitted")
			Expect(<-recorder.Events).To(ContainSubstring(str))
			resetNode()
		}
	})

	It("should return an event for empty mandatory annotation", func() {
		for annotation, str := range map[string]string{
			AnnotationNodeAvailabilityZone: "EmptyAvailabilityZone",
			AnnotationNodeCapacity:         "EmptyCapacity",
			AnnotationNodeType:             "EmptyType",
		} {
			By(annotation)
			annotations := node.GetAnnotations()
			annotations[annotation] = ""
			node.SetAnnotations(annotations)
			_, err = provider.GetNodeInfo(ctx, recorder, node)
			ExpectWithOffset(1, err).ToNot(HaveOccurred())
			ExpectWithOffset(2, recorder.Events).To(HaveLen(1), "no event emitted")
			Expect(<-recorder.Events).To(ContainSubstring(str))
			resetNode()
		}
	})

	It("should return an error for broken selector", func() {
		for _, annotation := range []string{
			AnnotationNodeAvailabilityZone,
			AnnotationNodeCapacity,
			AnnotationNodeType,
		} {
			annotations := node.GetAnnotations()
			annotations[annotation] = "annotation=absent"
			node.SetAnnotations(annotations)
			_, err = provider.GetNodeInfo(ctx, recorder, node)
			ExpectWithOffset(1, err).To(HaveOccurred())
			if len(recorder.Events) > 0 {
				ExpectWithOffset(3, recorder.Events).To(BeEmpty(),
					fmt.Sprintf("unexpected event is emitted: %s", <-recorder.Events))
			}
			resetNode()
		}
	})
})
