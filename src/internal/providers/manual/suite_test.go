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
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provider manual")
}

var (
	cancel   context.CancelFunc
	ctx      context.Context
	err      error
	provider Provider
	recorder *record.FakeRecorder
	scheme   *runtime.Scheme
)

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	scheme = runtime.NewScheme()
	corev1.AddToScheme(scheme)
	recorder = record.NewFakeRecorder(1)
	ctx, cancel = context.WithCancel(context.Background())
	provider = Provider{}
})

var _ = AfterSuite(func() {
	cancel()
})

// NewFakeNode creates a Node with a GVK and name only, no other fields
func NewFakeNode() (node *corev1.Node) {
	node = &corev1.Node{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}

	var gvk []schema.GroupVersionKind
	var unregistered bool
	gvk, unregistered, err = scheme.ObjectKinds(node)
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get object kinds for a Node")
	ExpectWithOffset(2, unregistered).To(BeFalse())
	ExpectWithOffset(3, gvk).To(HaveLen(1))
	node.APIVersion, node.Kind = gvk[0].ToAPIVersionAndKind()
	return
}
