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
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewTypeMeta(obj runtime.Object, scheme *runtime.Scheme) (m metav1.TypeMeta) {
	if scheme == nil {
		scheme = runtime.NewScheme()
		appsv1.AddToScheme(scheme)
		batchv1.AddToScheme(scheme)
		corev1.AddToScheme(scheme)
	}
	gvk, unregistered, err := scheme.ObjectKinds(obj)
	ExpectWithOffset(1, err).ToNot(HaveOccurred(), "failed to get object kinds for a Node")
	ExpectWithOffset(2, unregistered).To(BeFalse())
	ExpectWithOffset(3, gvk).To(HaveLen(1))
	m.APIVersion, m.Kind = gvk[0].ToAPIVersionAndKind()
	return
}
