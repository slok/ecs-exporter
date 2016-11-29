package collector

import (
	"testing"

	"github.com/golang/mock/gomock"
	awsMock "github.com/slok/ecs-exporter/mock/aws"
	"github.com/slok/ecs-exporter/mock/aws/sdk"
)

func TestGetClusters(t *testing.T) {
	tests := []struct {
		ids       []string
		wantError bool
	}{
		{
			ids:       []string{"cluster1", "cluster2", "cluster3", "cluster4"},
			wantError: false,
		},
		{
			ids:       []string{"cluster1"},
			wantError: false,
		},
		{
			ids:       []string{},
			wantError: false,
		},
		{
			ids:       []string{},
			wantError: true,
		},
	}

	for _, test := range tests {

		// Mock
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockECS := sdk.NewMockECSAPI(ctrl)
		awsMock.MockECSListClusters(t, mockECS, test.wantError, test.ids...)

		e := &ECSClient{
			client: mockECS,
		}

		cs, err := e.GetClusterIDs()
		if !test.wantError {
			if err != nil {
				t.Errorf("\n- %v\n-  Shouldn't return an error, it did: %v", test, err)
			}

			if len(cs) != len(test.ids) {
				t.Errorf("\n- %v\n-  returned cluster lenght is wrong, want: %d; got: %d", test, len(test.ids), len(cs))
			}

			for i := 0; i < len(cs); i++ {
				if cs[i] != test.ids[i] {
					t.Errorf("\n- %v\n-  Id in cluster is wrong, want: %s; got: %s", test, test.ids[i], cs[i])
				}
			}

		} else {
			if err == nil {
				t.Errorf("\n- %v\n-  Should return an error, it didn't", test)
			}
		}

	}
}
