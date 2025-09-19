package types

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
