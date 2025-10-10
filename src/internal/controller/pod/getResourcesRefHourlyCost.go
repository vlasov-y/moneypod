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

package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetResourcesRefHourlyCost calculates the hourly cost per CPU core and memory MiB for a node
func GetResourcesRefHourlyCost(
	ctx context.Context, c client.Client, r record.EventRecorder,
	pod *corev1.Pod, node *corev1.Node, nodeHourlyCost float64) (cpuCoreCost float64, memoryMiBCost float64, err error) {

	// Define base resource units
	cpuCore := resource.MustParse("1.0")
	memoryGiB := resource.MustParse("1Gi")
	cpuCoreFloat := cpuCore.AsApproximateFloat64()
	memoryGiBFloat := memoryGiB.AsApproximateFloat64()

	// Get node's allocatable resources
	allocatableCPU := node.Status.Allocatable.Cpu().AsApproximateFloat64()
	allocatableMemory := node.Status.Allocatable.Memory().AsApproximateFloat64()

	// Calculate total resource units and cost per unit
	// We sum count of Cores and GiBs and divide hourly cost on that value
	// So for example you have 2.0/8Gi, so there is 2 cores + 8 Gi = 10 units
	// Hourly price is 0.035, so 1 core == 1 Gi == 0.0035
	unitsCount := allocatableCPU/cpuCoreFloat + allocatableMemory/memoryGiBFloat
	unitHourlyCost := nodeHourlyCost / unitsCount

	// Set costs per resource type
	cpuCoreCost = unitHourlyCost
	memoryMiBCost = unitHourlyCost / 1024 // Convert from GiB to MiB

	return
}
