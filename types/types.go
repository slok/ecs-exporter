package types

const (
	ContainerInstanceStatusActive   = "ACTIVE"
	ContainerInstanceStatusInactive = "INACTIVE"
)

// ECSService represents a service on an ECS cluster
type ECSService struct {
	ID                           string // Service ARN
	Name                         string // Name of the service
	DesiredT, PendingT, RunningT int64  // Service task information
	Deployments                  int    // Number of deployments
}

// ECSCluster reprensens a cluster on ECS
type ECSCluster struct {
	ID   string // Cluster ARN
	Name string // Name of the service
}

// ECSContainerInstance represents a cluster container instance
type ECSContainerInstance struct {
	ID         string // Container instance ARN
	InstanceID string // EC2 instance ID
	AgentConn  bool   // The state of container instnace agent
	Active     bool   // The state of the container instance
	PendingT   int64  // The number of tasks in the container instance with pending state
}
