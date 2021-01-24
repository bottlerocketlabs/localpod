package docker

import (
	"reflect"
	"testing"
)

func TestEnvToArgs(t *testing.T) {
	env := map[string]string{
		"KEY1": "val1",
		"KEY2": "val2",
	}
	expected := []string{"--env", "KEY1=\"val1\"", "--env", "KEY2=\"val2\""}
	args := envToDockerArgs(env)
	if !reflect.DeepEqual(expected, args) {
		t.Errorf("got: %v, expected: %v", args, expected)
	}
}
