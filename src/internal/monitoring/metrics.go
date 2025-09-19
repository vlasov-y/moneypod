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

	PodHourlyCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "pod",
		Name:      "hourly_cost",
		Help:      "Pod hourly cost.",
	}, []string{"name", "namespace", "owner_kind", "owner_name"})
	PodTotalCostMetric = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Subsystem: "pod",
		Name:      "total_cost",
		Help:      "Pod total cost.",
	}, []string{"name", "namespace", "owner_kind", "owner_name"})
)

// RegisterMetrics registers all metrics in the Metrics map with Prometheus's global registry.
func RegisterMetrics() {
	metrics.Registry.MustRegister(NodeHourlyCostMetric)
	metrics.Registry.MustRegister(NodeTotalCostMetric)
	metrics.Registry.MustRegister(PodHourlyCostMetric)
	metrics.Registry.MustRegister(PodTotalCostMetric)
}
