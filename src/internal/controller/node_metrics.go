package controller

import (
	"math"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/vlasov-y/moneypod/internal/monitoring"
	"github.com/vlasov-y/moneypod/internal/types"
	corev1 "k8s.io/api/core/v1"
)

func hasExistingNodeMetrics(node *corev1.Node) bool {
	// Check if metrics with this node name already exist
	metricFamilies, _ := prometheus.DefaultGatherer.Gather()
	for _, mf := range metricFamilies {
		if mf.GetName() == "node_hourly_cost" {
			for _, metric := range mf.GetMetric() {
				for _, label := range metric.GetLabel() {
					if label.GetName() == "name" && label.GetValue() == node.Name {
						return true
					}
				}
			}
		}
	}
	return false
}

func updateNodeMetrics(node *corev1.Node, cost float64) {
	// Update only the cost values for existing metrics with this node name
	metricFamilies, _ := prometheus.DefaultGatherer.Gather()
	for _, mf := range metricFamilies {
		if mf.GetName() == "node_hourly_cost" {
			for _, metric := range mf.GetMetric() {
				var nodeName string
				var labels []string
				for _, label := range metric.GetLabel() {
					if label.GetName() == "name" {
						nodeName = label.GetValue()
					}
					labels = append(labels, label.GetValue())
				}
				// Find metrics labels for the node and use for updating the values
				if nodeName == node.Name {
					monitoring.NodeHourlyCostMetric.WithLabelValues(labels...).Set(cost)
					hours := math.Ceil(time.Since(node.GetCreationTimestamp().Time).Hours())
					monitoring.NodeTotalCostMetric.WithLabelValues(labels...).Set(cost * hours)
				}
			}
		}
	}
}

func deleteNodeMetrics(node *corev1.Node) {
	monitoring.NodeHourlyCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": node.Name,
	})
	monitoring.NodeTotalCostMetric.DeletePartialMatch(prometheus.Labels{
		"name": node.Name,
	})
}

func createNodeMetrics(node *corev1.Node, cost float64, info *types.NodeInfo) {
	monitoring.NodeHourlyCostMetric.WithLabelValues(
		node.Name, info.Type, string(info.Capacity),
		info.Id, info.AvailabilityZone,
	).Set(cost)
	// Calculate hours passed since creation
	hours := math.Ceil(time.Since(node.GetCreationTimestamp().Time).Hours())
	monitoring.NodeTotalCostMetric.WithLabelValues(
		node.Name, info.Type, string(info.Capacity),
		info.Id, info.AvailabilityZone,
	).Set(cost * hours)
}
