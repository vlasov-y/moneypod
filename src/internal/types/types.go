package types

const (
	// Node hourly cost
	AnnotationHourlyCost = "moneypod.io/hourly-cost"
	// Hourly cost copied from the node where the pod runs
	AnnotationNodeHourlyCost = "moneypod.io/node-hourly-cost"
	// Max concurrect reconciles per controller
	MaxConcurrentReconciles = 1
)

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
	Capacity NodeCapacity
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
}
