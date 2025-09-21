package controller

import (
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vlasov-y/moneypod/internal/monitoring"
	"github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
)

func deleteNodeMetrics(node *corev1.Node) {
	monitoring.NodeHourlyCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": node.Name,
	})
	monitoring.NodeTotalCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": node.Name,
	})
}

func createNodeMetrics(node *corev1.Node, cost float64, info *types.NodeInfo) {
	deleteNodeMetrics(node)
	monitoring.NodeHourlyCostMetric.WithLabelValues(
		node.Name, info.Type, info.Capacity,
		info.Id, info.AvailabilityZone,
	).Set(cost)
	// Calculate hours passed since creation
	hours := math.Ceil(time.Since(node.GetCreationTimestamp().Time).Hours())
	monitoring.NodeTotalCostMetric.WithLabelValues(
		node.Name, info.Type, info.Capacity,
		info.Id, info.AvailabilityZone,
	).Set(cost * hours)
}
