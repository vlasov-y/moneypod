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

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NewFakeDeployment", Ordered, func() {
	Context("when making an object", func() {
		It("should return a valid deployment", func() {
			deployment := NewFakeDeployment()
			ExpectWithOffset(1, deployment.Kind).To(Equal("Deployment"))
			ExpectWithOffset(2, deployment.Name).ToNot(BeEmpty())
			ExpectWithOffset(3, deployment.Status).ToNot(BeNil())
		})
	})
})
