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
package utils

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils")
}

var _ = Describe("Utils", Ordered, func() {
	Context("when checking requeue error", func() {
		It("should return true on requeue error", func() {
			Expect(CheckRequeue(ErrRequestRequeue)).To(BeTrue())
		})

		It("should return false for any other error", func() {
			Expect(CheckRequeue(nil)).To(BeFalse())
			Expect(CheckRequeue(errors.New(""))).To(BeFalse())
			Expect(CheckRequeue(errors.New("value"))).To(BeFalse())
		})
	})
})
