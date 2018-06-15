package resolver

// EventType is a type of event
type EventType string

const (
	// Added is added
	Added EventType = "ADDED"
	// Modified is modified
	Modified EventType = "MODIFIED"
	// Deleted is deleted
	Deleted EventType = "DELETED"
	// Error is error
	Error EventType = "ERROR"
)

// Event represents a single event to a watched resource.
type Event struct {
	Type   EventType `json:"type"`
	Object Endpoints `json:"object"`
}

// Endpoints are things to connect to
type Endpoints struct {
	Kind       string   `json:"kind"`
	APIVersion string   `json:"apiVersion"`
	Metadata   Metadata `json:"metadata"`
	Subsets    []Subset `json:"subsets"`
}

// Metadata is things about things
type Metadata struct {
	Name            string `json:"name"`
	ResourceVersion string `json:"resourceVersion"`
}

// Subset is a bit of a superset
type Subset struct {
	Addresses []Address `json:"addresses"`
	Ports     []Port    `json:"ports"`
}

// Address is an IP
type Address struct {
	IP string `json:"ip"`
}

// Port is a port
type Port struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}
