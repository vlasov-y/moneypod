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

	. "github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

func (provider *Provider) GetNodeInfo(ctx context.Context, r record.EventRecorder, node *corev1.Node) (info NodeInfo, err error) {
	var exists bool
	annotations := node.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	info.ID = "manual"

	// Get node's capacity from labels/annotations
	if info.Capacity, exists = annotations[AnnotationNodeCapacity]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoCapacity", fmt.Sprintf("%s is not defined", AnnotationNodeCapacity))
		return
	}
	if info.Capacity, err = provider.parseAnnotationLabelSelector(node, info.Capacity); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "CapacityGetError", err.Error())
		return
	}
	if info.Capacity == "" {
		r.Eventf(node, corev1.EventTypeWarning, "EmptyCapacity", fmt.Sprintf("%s is an empty string", AnnotationNodeCapacity))
		return
	}

	// Get node's type from labels/annotations
	if info.Type, exists = annotations[AnnotationNodeType]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoType", fmt.Sprintf("%s is not defined", AnnotationNodeType))
		return
	}
	if info.Type, err = provider.parseAnnotationLabelSelector(node, info.Type); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "TypeGetError", err.Error())
		return
	}
	if info.Type == "" {
		r.Eventf(node, corev1.EventTypeWarning, "EmptyType", fmt.Sprintf("%s is an empty string", AnnotationNodeType))
		return
	}

	// Get node's availability zone from labels/annotations
	if info.AvailabilityZone, exists = annotations[AnnotationNodeAvailabilityZone]; !exists {
		r.Eventf(node, corev1.EventTypeWarning, "NoAvailabilityZone", fmt.Sprintf("%s is not defined", AnnotationNodeAvailabilityZone))
		return
	}
	if info.AvailabilityZone, err = provider.parseAnnotationLabelSelector(node, info.AvailabilityZone); err != nil {
		r.Eventf(node, corev1.EventTypeWarning, "AvailabilityZoneGetError", err.Error())
		return
	}
	if info.AvailabilityZone == "" {
		r.Eventf(node, corev1.EventTypeWarning, "EmptyAvailabilityZone", fmt.Sprintf("%s is an empty string", AnnotationNodeAvailabilityZone))
		return
	}

	return
}
