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
	monitoring.PodHourlyCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": pod.Name,
	})
	monitoring.PodTotalCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": pod.Name,
	})
}

func updatePodMetrics(pod *corev1.Pod, cost float64, info *types.PodInfo) {
	monitoring.PodHourlyCostMetric.WithLabelValues(
		pod.Name, pod.Namespace, info.Owner.Kind, info.Owner.Name, pod.Spec.NodeName,
	).Set(cost)
	// Calculate hours passed since creation
	hours := math.Ceil(time.Since(pod.GetCreationTimestamp().Time).Hours())
	monitoring.PodTotalCostMetric.WithLabelValues(
		pod.Name, pod.Namespace, info.Owner.Kind, info.Owner.Name, pod.Spec.NodeName,
	).Set(cost * hours)
}
