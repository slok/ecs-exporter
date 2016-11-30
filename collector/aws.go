package collector

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"

	"github.com/slok/ecs-exporter/types"
)

// Generate ECS API mocks running go generate
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/ecs/ecsiface/interface.go -package sdk -destination ../mock/aws/sdk/ecsiface_mock.go

// ECSClient is a wrapper for AWS ecs client that implements helpers to get ECS clusters metrics
type ECSClient struct {
	client        ecsiface.ECSAPI
	apiMaxResults int64
}

// NewECSClient will return an initialized ECSClient
func NewECSClient(awsRegion string) (*ECSClient, error) {
	// Create AWS session
	s := session.New(&aws.Config{Region: aws.String(awsRegion)})
	if s == nil {
		return nil, fmt.Errorf("error creating aws session")
	}

	return &ECSClient{
		client:        ecs.New(s),
		apiMaxResults: 100,
	}, nil
}

// GetClusters will get the clusters from the ECS API
func (e *ECSClient) GetClusters() ([]*types.ECSCluster, error) {
	cArns := []*string{}
	params := &ecs.ListClustersInput{
		MaxResults: aws.Int64(e.apiMaxResults),
	}

	// Get cluster IDs
	for {
		resp, err := e.client.ListClusters(params)
		if err != nil {
			return nil, err
		}

		for _, c := range resp.ClusterArns {
			cArns = append(cArns, c)
		}
		if resp.NextToken == nil || aws.StringValue(resp.NextToken) == "" {
			break
		}
		params.NextToken = resp.NextToken
	}

	// Get service descriptions
	params2 := &ecs.DescribeClustersInput{
		Clusters: cArns,
	}
	resp2, err := e.client.DescribeClusters(params2)
	if err != nil {
		return nil, err
	}

	cs := []*types.ECSCluster{}
	for _, c := range resp2.Clusters {
		ec := &types.ECSCluster{
			ID:   aws.StringValue(c.ClusterArn),
			Name: aws.StringValue(c.ClusterName),
		}
		cs = append(cs, ec)
	}

	return cs, nil
}

// GetClusterServices will return all the services from a cluster
func (e *ECSClient) GetClusterServices(clusterArn string) ([]*types.ECSService, error) {
	sArns := []*string{}

	// Get service ids
	params := &ecs.ListServicesInput{
		Cluster:    aws.String(clusterArn),
		MaxResults: aws.Int64(e.apiMaxResults),
	}

	for {
		resp, err := e.client.ListServices(params)
		if err != nil {
			return nil, err
		}

		for _, s := range resp.ServiceArns {
			sArns = append(sArns, s)
		}

		if resp.NextToken == nil || aws.StringValue(resp.NextToken) == "" {
			break
		}
		params.NextToken = resp.NextToken
	}

	// Get service descriptions
	params2 := &ecs.DescribeServicesInput{
		Services: sArns,
	}
	resp2, err := e.client.DescribeServices(params2)
	if err != nil {
		return nil, err
	}

	ss := []*types.ECSService{}
	for _, s := range resp2.Services {
		es := &types.ECSService{
			ID:       aws.StringValue(s.ServiceArn),
			Name:     aws.StringValue(s.ServiceName),
			DesiredT: aws.Int64Value(s.DesiredCount),
			RunningT: aws.Int64Value(s.RunningCount),
			PendingT: aws.Int64Value(s.PendingCount),
		}
		ss = append(ss, es)
	}

	return ss, nil
}
