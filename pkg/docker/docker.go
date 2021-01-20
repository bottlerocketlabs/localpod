package docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/stuart-warren/localpod/pkg/config"
)

const (
	DefaultContainerEntrypoint = "/bin/sh"
	DefaultContainerCommand    = "echo Container started ; trap \"exit 0\" 15; while sleep 1 & wait $!; do :; done"
	DefaultContainerHostname   = "localpod"
)

func HasDocker() bool {
	_, err := exec.LookPath("docker")
	if err != nil {
		return false
	}
	return true
}

type Container struct {
	ID     string
	Name   string
	Config *config.DevContainer
	Args   []string
}

func envToDockerArgs(env map[string]string) []string {
	var args []string
	for k, v := range env {
		args = append(args, "--env")
		args = append(args, fmt.Sprintf("%s=%q", k, v))
	}
	return args
}

func mountsToDockerArgs(mounts []string) []string {
	var args []string
	for _, m := range mounts {
		args = append(args, "--mount")
		args = append(args, fmt.Sprintf("%q", m))
	}
	return args
}

func expandEnvArgs(args []string, env config.Env) []string {
	var out []string
	for _, e := range args {
		out = append(out, os.Expand(e, env.Get))
	}
	return out
}

func buildArgs(command string, name string, cfg *config.DevContainer) []string {
	var args []string
	args = append(args, command)
	args = append(args, "--tty", "--interactive")
	args = append(args, "--name", name)
	args = append(args, "--hostname", DefaultContainerHostname)
	args = append(args, "--user", cfg.ContainerUser)
	args = append(args, envToDockerArgs(cfg.ContainerEnv)...)
	args = append(args, mountsToDockerArgs(cfg.Mounts)...)
	if cfg.WorkspaceMount != "" {
		args = append(args, "--mount", cfg.WorkspaceMount)
	}
	args = append(args, cfg.RunArgs...)
	if cfg.OverrideCommand {
		args = append(args, "--entrypoint", DefaultContainerEntrypoint)
	}
	args = append(args, cfg.Image) // Image
	if cfg.OverrideCommand {
		args = append(args, "-c", DefaultContainerCommand)
	}
	return args
}

func CreateContainer(name string, env config.Env, cfg *config.DevContainer) (Container, error) {
	c := Container{
		Args:   expandEnvArgs(buildArgs("create", name, cfg), env),
		Config: cfg,
		Name:   name,
	}
	fmt.Printf("DEBUG: args: %v\n", c.Args)
	out, err := exec.Command("docker", c.Args...).Output()
	if err != nil {
		return c, fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	c.ID = string(out)
	return c, nil
}

func (c *Container) Start() error {
	_, err := exec.Command("docker", "start", c.Name).Output()
	if err != nil {
		return fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	return nil
}

func (c *Container) Exec(stdin io.Reader, stdout, stderr io.Writer) error {
	args := []string{"exec", "--tty", "--interactive"}
	args = append(args, c.Config.ExecCommand...)
	cmd := exec.Command("docker", args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	cmd.Stdin = stdin
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	return nil
}
