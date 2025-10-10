// Copyright 2025 The MoneyPod Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package types defines common data structures and constants used throughout the application.
package types

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
)

const (
	annotationDomain = "moneypod.io"
	// Node hourly cost
	AnnotationNodeHourlyCost = annotationDomain + "/node-hourly-cost"
	// Spot or on-demand
	AnnotationNodeCapacity = annotationDomain + "/capacity"
	// Instance type: t3a.small, etc.
	AnnotationNodeType = annotationDomain + "/type"
	// Node location
	AnnotationNodeAvailabilityZone = annotationDomain + "/availability-zone"
	// Placeholder for an unknown price
	UnknownCost = "unknown"
)

type Reconciler struct {
	Config                  *rest.Config
	Scheme                  *runtime.Scheme
	Recorder                record.EventRecorder
	MaxConcurrentReconciles int
}

type NodeCapacity string

const (
	Spot     NodeCapacity = "spot"
	OnDemand NodeCapacity = "on-demand"
)

// NodeInfo contains provider information about the node.
type NodeInfo struct {
	// Provider node ID
	ID string
	// Instance type: t3a.small, etc.
	Type string
	// Spot or on-demand
	Capacity string
	// Availability zone
	AvailabilityZone string
}

// PodInfo contains provider information about the pod.
type PodInfo struct {
	// Pod owner reference
	Owner struct {
		Kind string
		Name string
	}
	NodeHourlyCost          float64
	NodeCPUCoreHourlyCost   float64
	NodeMemoryMiBHourlyCost float64
	PodRequestsHourlyCost   float64
}
