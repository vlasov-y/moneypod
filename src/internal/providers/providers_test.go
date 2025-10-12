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

// Package providers provides an interface to implement by all cloud proviers for getting costs and information about the nodes.
package providers

import (
	"reflect"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/vlasov-y/moneypod/internal/providers/aws"
	"github.com/vlasov-y/moneypod/internal/providers/manual"
	corev1 "k8s.io/api/core/v1"
)

func TestProviders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Providers")
}

var _ = Describe("NodeReconciler", Ordered, func() {
	Context("when creating a provider", func() {
		It("should return AWS provider for aws:// provider ID prefix", func() {
			provider := NewProvider(&corev1.Node{
				Spec: corev1.NodeSpec{ProviderID: "aws:///eu-central-1a/i-abcdef1234"},
			})
			Expect(reflect.TypeOf(provider)).To(Equal(reflect.TypeFor[*aws.Provider]()))
		})

		It("should return manual provider for an unmatched prefix", func() {
			provider := NewProvider(&corev1.Node{
				Spec: corev1.NodeSpec{ProviderID: "something"},
			})
			Expect(reflect.TypeOf(provider)).To(Equal(reflect.TypeFor[*manual.Provider]()))
		})

		It("should return manual provider for an absent provider ID", func() {
			provider := NewProvider(&corev1.Node{})
			Expect(reflect.TypeOf(provider)).To(Equal(reflect.TypeFor[*manual.Provider]()))
		})
	})
})
