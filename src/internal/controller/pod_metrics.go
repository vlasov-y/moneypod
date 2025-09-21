package controller

import (
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vlasov-y/moneypod/internal/monitoring"
	"github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
)

func deletePodMetrics(pod *corev1.Pod) {
	monitoring.PodCpuHourlyCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": pod.Name, "namespace": pod.Namespace,
	})
	monitoring.PodMemoryHourlyCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": pod.Name, "namespace": pod.Namespace,
	})
	monitoring.PodRequestsTotalCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": pod.Name, "namespace": pod.Namespace,
	})
}

func createPodMetrics(pod *corev1.Pod, node *corev1.Node, info *types.PodInfo) {
	deletePodMetrics(pod)
	monitoring.PodCpuHourlyCostMetric.WithLabelValues(
		pod.Name, pod.Namespace, info.Owner.Kind, info.Owner.Name, pod.Spec.NodeName,
	).Set(info.NodeCpuCoreHourlyCost)
	monitoring.PodMemoryHourlyCostMetric.WithLabelValues(
		pod.Name, pod.Namespace, info.Owner.Kind, info.Owner.Name, pod.Spec.NodeName,
	).Set(info.NodeMemoryMiBHourlyCost)
	hours := math.Ceil(time.Since(pod.GetCreationTimestamp().Time).Hours())
	monitoring.PodRequestsTotalCostMetric.WithLabelValues(
		pod.Name, pod.Namespace, info.Owner.Kind, info.Owner.Name, pod.Spec.NodeName,
	).Set(info.PodRequestsHourlyCost * hours)
}
