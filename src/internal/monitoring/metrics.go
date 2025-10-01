/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package monitoring provides metrics collection and monitoring functionality.
package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

// Metrics is a map to store created metrics.
var (
	metricsNamespace = "moneypod"

	NodeHourlyCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "node",
		Name:      "hourly_cost",
		Help:      "Node hourly cost.",
	}, []string{"node", "name", "type", "capacity", "id", "availability_zone"})

	PodCPUHourlyCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "pod",
		Name:      "cpu_hourly_cost",
		Help:      "Pod CPU hourly cost for one CPU core.",
	}, []string{"pod", "name", "namespace", "owner_kind", "owner_name", "node"})
	PodMemoryHourlyCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "pod",
		Name:      "memory_hourly_cost",
		Help:      "Pod Memory hourly cost for one MiB.",
	}, []string{"pod", "name", "namespace", "owner_kind", "owner_name", "node"})
	PodRequestsHourlyCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "pod",
		Name:      "requests_hourly_cost",
		Help:      "Pod resources requests hourly cost.",
	}, []string{"pod", "name", "namespace", "owner_kind", "owner_name", "node"})
)

// RegisterMetrics registers all metrics in the Metrics map with Prometheus's global registry.
func RegisterMetrics() {
	metrics.Registry.MustRegister(NodeHourlyCostMetric)
	metrics.Registry.MustRegister(PodCPUHourlyCostMetric)
	metrics.Registry.MustRegister(PodMemoryHourlyCostMetric)
	metrics.Registry.MustRegister(PodRequestsHourlyCostMetric)
}
