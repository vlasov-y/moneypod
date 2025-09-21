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
	}, []string{"name", "type", "capacity", "id", "availability_zone"})
	NodeTotalCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "node",
		Name:      "total_cost",
		Help:      "Node total cost.",
	}, []string{"name", "type", "capacity", "id", "availability_zone"})

	PodCpuHourlyCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "pod",
		Name:      "cpu_hourly_cost",
		Help:      "Pod CPU hourly cost for one CPU core.",
	}, []string{"name", "namespace", "owner_kind", "owner_name", "node"})
	PodMemoryHourlyCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "pod",
		Name:      "memory_hourly_cost",
		Help:      "Pod Memory hourly cost for one MiB.",
	}, []string{"name", "namespace", "owner_kind", "owner_name", "node"})
	PodRequestsTotalCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "pod",
		Name:      "requests_total_cost",
		Help:      "Pod resources requests total cost.",
	}, []string{"name", "namespace", "owner_kind", "owner_name", "node"})
)

// RegisterMetrics registers all metrics in the Metrics map with Prometheus's global registry.
func RegisterMetrics() {
	metrics.Registry.MustRegister(NodeHourlyCostMetric)
	metrics.Registry.MustRegister(NodeTotalCostMetric)
	metrics.Registry.MustRegister(PodCpuHourlyCostMetric)
	metrics.Registry.MustRegister(PodMemoryHourlyCostMetric)
	metrics.Registry.MustRegister(PodRequestsTotalCostMetric)
}
