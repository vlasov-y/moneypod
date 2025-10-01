package controller

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/vlasov-y/moneypod/internal/monitoring"
	"github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
)

func deletePodMetrics(pod *corev1.Pod) {
	monitoring.PodCPUHourlyCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": pod.Name, "namespace": pod.Namespace,
	})
	monitoring.PodMemoryHourlyCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": pod.Name, "namespace": pod.Namespace,
	})
	monitoring.PodRequestsHourlyCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": pod.Name, "namespace": pod.Namespace,
	})
}

func createPodMetrics(pod *corev1.Pod, info *types.PodInfo) {
	deletePodMetrics(pod)
	monitoring.PodCPUHourlyCostMetric.WithLabelValues(
		pod.Name, pod.Name, pod.Namespace, info.Owner.Kind, info.Owner.Name, pod.Spec.NodeName,
	).Set(info.NodeCPUCoreHourlyCost)
	monitoring.PodMemoryHourlyCostMetric.WithLabelValues(
		pod.Name, pod.Name, pod.Namespace, info.Owner.Kind, info.Owner.Name, pod.Spec.NodeName,
	).Set(info.NodeMemoryMiBHourlyCost)
	monitoring.PodRequestsHourlyCostMetric.WithLabelValues(
		pod.Name, pod.Name, pod.Namespace, info.Owner.Kind, info.Owner.Name, pod.Spec.NodeName,
	).Set(info.PodRequestsHourlyCost)
}
