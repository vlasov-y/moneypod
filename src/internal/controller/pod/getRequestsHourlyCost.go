package pod

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func GetRequestsHourlyCost(
	ctx context.Context, c client.Client, r record.EventRecorder,
	pod *corev1.Pod, node *corev1.Node, nodeHourlyCost float64) (hourlyCost float64, err error) {
	log := logf.FromContext(ctx)

	allocatedCPU := resource.Quantity{}
	allocatedMemory := resource.Quantity{}

	// Sum allocated resources from pod status
	for _, status := range pod.Status.ContainerStatuses {
		if status.AllocatedResources.Cpu() != nil {
			allocatedCPU.Add(*status.AllocatedResources.Cpu())
		}
		if status.AllocatedResources.Memory() != nil {
			allocatedMemory.Add(*status.AllocatedResources.Memory())
		}
	}

	// Get reference cost
	var cpuCoreCost, memoryMiBCost float64
	cpuCoreCost, memoryMiBCost, err = GetResourcesRefHourlyCost(ctx, c, r, pod, node, nodeHourlyCost)
	if err != nil {
		return
	}

	// Define base resource units
	cpuCore := resource.MustParse("1.0")
	memoryMiB := resource.MustParse("1Mi")
	cpuCoreFloat := cpuCore.AsApproximateFloat64()
	memoryMiBFloat := memoryMiB.AsApproximateFloat64()

	// Calculate resources requests cost
	cpuCost := allocatedCPU.AsApproximateFloat64() / cpuCoreFloat * cpuCoreCost
	memoryCost := allocatedMemory.AsApproximateFloat64() / memoryMiBFloat * memoryMiBCost
	hourlyCost = cpuCost + memoryCost
	log.V(2).Info("pod requests hourly cost", "cpu", cpuCost, "memory", memoryCost, "sum", hourlyCost)

	return
}
