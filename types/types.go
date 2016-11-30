package types

// ECSService represents a service on an ECS cluster
type ECSService struct {
	ID                           string // Service ARN
	Name                         string // Name of the service
	DesiredT, PendingT, RunningT int64  // Service task information
}

// ECSCluster reprensens a cluster on ECS
type ECSCluster struct {
	ID   string // Cluster ARN
	Name string // Name of the service
}
