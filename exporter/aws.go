package exporter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
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

// GetClusterIDs will get the clusters from the ECS API
func (e *ECSClient) GetClusterIDs() ([]*string, error) {
	clusters := []*string{}
	params := &ecs.ListClustersInput{
		MaxResults: aws.Int64(e.apiMaxResults),
	}

	// Get cluster IDs
	for {
		resp, err := e.client.ListClusters(params)
		if err != nil {
			return clusters, err
		}

		for _, c := range resp.ClusterArns {
			clusters = append(clusters, c)
		}
		if resp.NextToken == nil || aws.StringValue(resp.NextToken) == "" {
			break
		}
		params.NextToken = resp.NextToken
	}
	return clusters, nil
}
