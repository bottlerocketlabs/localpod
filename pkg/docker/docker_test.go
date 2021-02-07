package docker

import (
	"strings"
	"testing"
)

func TestEnvToArgs(t *testing.T) {
	env := map[string]string{
		"KEY1": "val1",
		"KEY2": "val2",
	}
	expected := []string{"--env", "=", "--env", "="}
	args := envToDockerArgs(env)
	if len(args) != len(expected) {
		t.Errorf("got a length of %d, expected: %d", len(args), len(expected))
	}
	for i, v := range expected {
		if !strings.Contains(args[i], v) {
			t.Errorf("got: %v, expected: %v", args, expected)
			t.Fail()
		}
	}
}
