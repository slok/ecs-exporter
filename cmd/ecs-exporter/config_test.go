package main

import (
	"testing"
)

func TestConfigParse(t *testing.T) {
	tests := []struct {
		ok  bool
		cmd []string
	}{
		{true, []string{"--aws.region", "eu-west-1", "--web.listen-address", "0.0.0.0:9999", "--web.telemetry-path", "/metrics2", "--metrics.disable-cinstances"}},
		{true, []string{"--aws.region", "eu-west-1", "--web.telemetry-path", "/metrics2"}},
		{true, []string{"--aws.region", "eu-west-1", "--web.listen-address", "0.0.0.0:9999"}},
		{true, []string{"--aws.region", "eu-west-1"}},
		{true, []string{"--aws.region", "eu-west-1", "--debug"}},
		{true, []string{"--aws.region", "eu-west-1", "--aws.cluster-filter", ".*-prod-.*"}},
		{false, []string{"--aws.region", "eu-west-1", "--aws.cluster-filter", "["}},
		{false, []string{"--web.listen-address", "0.0.0.0:9999", "--web.telemetry-path", "/metrics2"}},

		{false, []string{}},
		{false, []string{"--aws-region", "eu-west-1"}},
	}

	for _, test := range tests {
		c := new()
		err := c.parse(test.cmd)
		if err != nil && test.ok {
			t.Errorf("\n- %v\n- Cmd parsing shoudn't fail, it did: %v", test, err)
		}

		if err == nil && !test.ok {
			t.Errorf("\n- %v\n- Cmd parsing shoud fail, it didn't", test)
		}
	}
}
