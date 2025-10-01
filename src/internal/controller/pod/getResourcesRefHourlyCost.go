package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetResourcesRefHourlyCost(
	ctx context.Context, c client.Client, r record.EventRecorder,
	pod *corev1.Pod, node *corev1.Node, nodeHourlyCost float64) (cpuCoreCost float64, memoryMiBCost float64, err error) {

	cpuCore := resource.MustParse("1.0")
	memoryMiB := resource.MustParse("1Mi")
	cpuCoreFloat := cpuCore.AsApproximateFloat64()
	memoryMiBFloat := memoryMiB.AsApproximateFloat64()

	allocatableCPU := node.Status.Allocatable.Cpu().AsApproximateFloat64()
	allocatableMemory := node.Status.Allocatable.Memory().AsApproximateFloat64()

	cpuCoreCost = cpuCoreFloat * (nodeHourlyCost / 2) / allocatableCPU
	memoryMiBCost = memoryMiBFloat * (nodeHourlyCost / 2) / allocatableMemory

	return
}
