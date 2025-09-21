package types

import (
	"os"
	"strconv"
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
)

// Max concurrent reconciles per controller
var MaxConcurrentReconciles int = func() int {
	str := os.Getenv("MAX_CONCURRENT_RECONCILES")
	if str == "" {
		return 10
	}
	i, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		panic(err)
	}
	return int(i)
}()

type NodeCapacity string

const (
	Spot     NodeCapacity = "spot"
	OnDemand NodeCapacity = "on-demand"
)

// Provider information about the node
type NodeInfo struct {
	// Provider node ID
	Id string
	// Instance type: t3a.small, etc.
	Type string
	// Spot or on-demand
	Capacity string
	// Availability zone
	AvailabilityZone string
}

// Provider information about the node
type PodInfo struct {
	// Pod owner reference
	Owner struct {
		Kind string
		Name string
	}
	NodeHourlyCost          float64
	NodeCpuCoreHourlyCost   float64
	NodeMemoryMiBHourlyCost float64
	PodRequestsHourlyCost   float64
}
