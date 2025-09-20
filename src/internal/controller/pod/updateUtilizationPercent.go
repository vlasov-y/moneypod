package pod

import (
	"context"
	"errors"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metricsclientset "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func UpdateUtilizationPercent(ctx context.Context, rc *rest.Config, c client.Client, r record.EventRecorder, pod *corev1.Pod) (utilization float64, err error) {
	log := logf.FromContext(ctx)
	utilization = 0

	// Get metrics using direct client to avoid watch issues
	var podMetrics *metricsv1beta1.PodMetrics
	if metricsClient, err := metricsclientset.NewForConfig(rc); err == nil {
		if metrics, err := metricsClient.MetricsV1beta1().PodMetricses(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{}); err == nil {
			podMetrics = metrics
		}
	}

	// Calculate actual consumption and allocated resources
	actualCpu := resource.Quantity{}
	actualMemory := resource.Quantity{}
	allocatedCpu := resource.Quantity{}
	allocatedMemory := resource.Quantity{}

	// Sum actual consumption from metrics (if available)
	if podMetrics != nil {
		for _, container := range podMetrics.Containers {
			if container.Usage.Cpu() != nil {
				actualCpu.Add(*container.Usage.Cpu())
			}
			if container.Usage.Memory() != nil {
				actualMemory.Add(*container.Usage.Memory())
			}
		}
	}

	// Sum allocated resources from pod status
	for _, status := range pod.Status.ContainerStatuses {
		if status.AllocatedResources.Cpu() != nil {
			allocatedCpu.Add(*status.AllocatedResources.Cpu())
		}
		if status.AllocatedResources.Memory() != nil {
			allocatedMemory.Add(*status.AllocatedResources.Memory())
		}
	}

	// Use the higher value between actual consumption and allocated resources
	finalCpu := actualCpu
	if allocatedCpu.Cmp(actualCpu) > 0 {
		finalCpu = allocatedCpu
	}

	finalMemory := actualMemory
	if allocatedMemory.Cmp(actualMemory) > 0 {
		finalMemory = allocatedMemory
	}

	// Get Pod's Node resources
	node := corev1.Node{}
	if err = c.Get(ctx, types.NamespacedName{Name: pod.Spec.NodeName}, &node); err != nil {
		if !apierrors.IsNotFound(err) {
			log.V(1).Error(err, "cannot get the node")
		}
		return utilization, client.IgnoreNotFound(err)
	}

	// Calculate utilization
	nodeCpuAllocatable := float64(node.Status.Allocatable.Cpu().MilliValue())
	nodeMemoryAllocatable := float64(node.Status.Allocatable.Memory().MilliValue())
	if nodeCpuAllocatable == 0 || nodeMemoryAllocatable == 0 {
		err = errors.New("node allocatable CPU or Memory is 0")
		log.V(1).Error(err, err.Error(), "cpu", nodeCpuAllocatable, "memory", nodeMemoryAllocatable)
		return
	}

	cpuUtilization := float64(finalCpu.MilliValue()) / nodeCpuAllocatable
	memoryUtilization := float64(finalMemory.MilliValue()) / nodeMemoryAllocatable
	utilization = max(cpuUtilization, memoryUtilization)
	log.V(2).Info("pod utilization", "cpu", cpuUtilization, "memory", memoryUtilization)

	return
}
