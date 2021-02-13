package docker

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"text/template"

	"github.com/bottlerocketlabs/localpod/pkg/config"
)

const (
	DefaultContainerEntrypoint = "/bin/sh"
	DefaultContainerCommand    = "echo Container started ; trap \"exit 0\" 15; while sleep 1 & wait $!; do :; done"
	DefaultContainerHostname   = "localpod"
	configSHA1EnvKey           = "LOCALPOD_CONFIG_SHA1"
)

var (
	ErrExistsButDifferent = fmt.Errorf("container exists, but does not match configuration")
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
}

func envToDockerArgs(env map[string]string) []string {
	var args []string
	for k, v := range env {
		args = append(args, "--env")
		args = append(args, fmt.Sprintf("%s=%s", k, v))
	}
	return args
}

func mountsToDockerArgs(mounts []string) []string {
	var args []string
	for _, m := range mounts {
		args = append(args, "--mount")
		args = append(args, m)
	}
	return args
}

func buildArgsToDockerArgs(buildArgs map[string]string, env config.Env) []string {
	var args []string
	for k, v := range buildArgs {
		args = append(args, "--build-arg")
		args = append(args, fmt.Sprintf("%s=%s", k, os.Expand(v, env.Get)))
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

func buildCreateArgs(name string, cfg *config.DevContainer) []string {
	var args []string
	args = append(args, "create")
	args = append(args, "--tty", "--interactive")
	args = append(args, "--name", name)
	args = append(args, "--pull", "always")
	args = append(args, "--hostname", DefaultContainerHostname)
	args = append(args, "--user", cfg.ContainerUser)
	args = append(args, envToDockerArgs(cfg.ContainerEnv)...)
	args = append(args, envToDockerArgs(map[string]string{
		configSHA1EnvKey: cfg.SHA1(),
	})...)
	args = append(args, mountsToDockerArgs(cfg.Mounts)...)
	if cfg.WorkspaceMount != "" {
		args = append(args, "--mount", cfg.WorkspaceMount)
	}
	args = append(args, "--workdir", cfg.WorkspaceFolder)
	args = append(args, cfg.RunArgs...)
	if cfg.OverrideCommand {
		args = append(args, "--entrypoint", DefaultContainerEntrypoint)
	}
	image := cfg.Image
	if cfg.Build.Dockerfile != "" && cfg.Build.Target != "" {
		image = cfg.Build.Target //assume was built successfully
	}
	args = append(args, image) // Image
	if cfg.OverrideCommand {
		args = append(args, "-c", DefaultContainerCommand)
	}
	return args
}

func BuildImage(cfg *config.DevContainer, env config.Env, stdout, stderr io.Writer) error {
	if cfg.Build.Dockerfile == "" {
		return nil
	}
	if cfg.Build.Context == "" {
		return fmt.Errorf("build.dockerfile set, but not build.context")
	}
	if cfg.Build.Target == "" && cfg.Image == "" {
		return fmt.Errorf("build.dockerfile set, but not build.target or image")
	}
	buildTarget := cfg.Build.Target
	if buildTarget == "" {
		buildTarget = cfg.Image
	}
	_, err := os.Stat(cfg.Build.Dockerfile)
	if os.IsNotExist(err) {
		return fmt.Errorf("configured build.dockerfile does not exist: %w", err)
	}
	if err != nil {
		return fmt.Errorf("failed to check for existance of build.dockerfile")
	}
	args := []string{"build", "--pull", "--tag", buildTarget, "--file", cfg.Build.Dockerfile}
	args = append(args, buildArgsToDockerArgs(cfg.Build.Args, env)...)
	args = append(args, cfg.Build.Context)
	fmt.Printf("DEBUG: Building image with: %s\n", args)
	cmd := exec.Command("docker", args...)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	return nil
}

func CreateContainer(name string, env config.Env, cfg *config.DevContainer) (Container, error) {
	c := Container{
		Config: cfg,
		Name:   name,
	}
	existingID, err := c.Exists(name)
	c.ID = existingID
	if err == nil {
		fmt.Printf("DEBUG: container already exists\n")
		return c, nil
	}
	if err == ErrExistsButDifferent {
		fmt.Printf("DEBUG: container already exists, but doesnt match configuration. Rebuilding\n")
		err := c.Rm()
		if err != nil {
			return c, fmt.Errorf("failed to remove existing container: %w", err)
		}
	}
	fmt.Printf("DEBUG: inpect: %s\n", err)
	args := expandEnvArgs(buildCreateArgs(name, cfg), env)
	fmt.Printf("DEBUG: args for create: %v\n", args)
	out, err := exec.Command("docker", args...).Output()
	if err != nil {
		return c, fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	c.ID = string(out)
	return c, nil
}

type containerInspect struct {
	ID     string `json:"Id"`
	Config struct {
		Env []string `json:"Env"`
	} `json:"Config"`
}

func (c *Container) Exists(name string) (string, error) {
	out, err := exec.Command("docker", "inspect", name).Output()
	if err != nil {
		return "", fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	var inspect []containerInspect
	err = json.Unmarshal(out, &inspect)
	if err != nil {
		return "", fmt.Errorf("unmarshal failed: %w", err)
	}
	if len(inspect) != 1 {
		return "", fmt.Errorf("unexpected number of inspect entries: %d", len(inspect))
	}
	expectedSHA1 := c.Config.SHA1()
	inspectEnv := config.NewEnv(inspect[0].Config.Env)
	inspectSHA1 := inspectEnv.Get(configSHA1EnvKey)
	if inspectSHA1 != expectedSHA1 {
		fmt.Printf("DEBUG: expected: %s, inspected: %s\n", expectedSHA1, inspectSHA1)
		return inspect[0].ID, ErrExistsButDifferent
	}
	return inspect[0].ID, nil
}

var (
	setupScript = `#!/bin/sh
set -e
# ensure user is created
adduser --home /home/{{.Username}} --gecos '' --disabled-password -u {{ .UID }} {{.Username}} || true
usermod --uid {{ .UID }} {{.Username}} || true
# setup user for passwordless sudo
addgroup sudo || true
addgroup {{.Username}} sudo || true
mkdir -p /etc/sudoers.d
echo "{{.Username}} ALL=(ALL) NOPASSWD:ALL" > "/etc/sudoers.d/{{.Username}}"
if command -v apk; then
	apk add --no-cache sudo
fi
if command -v apt-get; then
	apt-get update && apt-get install -y --no-install-recommends sudo
fi
# create dir for homebrew
mkdir -p /home/linuxbrew/.linuxbrew/Homebrew
chown -R {{.Username}} /home/linuxbrew
`
	startScript = `#!/bin/sh
# run the users default login shell
$(awk -F: -v user="{{.Username}}" '$1 == user {print $NF}' /etc/passwd) --login
`
	startCommand = "/start"
)

type TemplateScriptParams struct {
	UID      string
	Username string
}

func (c *Container) AddScript(name string, dstPath string, templatedContent string) error {
	tmpl, err := template.New(name).Parse(templatedContent)
	if err != nil {
		return fmt.Errorf("could not parse %s template: %w", name, err)
	}
	tmp := os.TempDir()
	scriptPath := filepath.Join(tmp, name)
	f, err := os.Create(scriptPath)
	if err != nil {
		return fmt.Errorf("could not create temp file")
	}
	defer os.Remove(f.Name())
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("could not get current user details")
	}
	params := TemplateScriptParams{
		UID:      currentUser.Uid,
		Username: c.Config.RemoteUser,
	}
	err = tmpl.Execute(f, params)
	if err != nil {
		return fmt.Errorf("could not execute %s template: %w", name, err)
	}
	err = f.Sync()
	if err != nil {
		return fmt.Errorf("could not sync %s to disk: %w", name, err)
	}
	_, err = exec.Command("docker", "cp", f.Name(), c.Name+":"+dstPath).Output()
	if err != nil {
		return fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	args := []string{"chmod", "+x", dstPath}
	err = c.RunCommand(args)
	if err != nil {
		return fmt.Errorf("could make %s executable: %w", dstPath, err)
	}
	return nil
}

func (c *Container) RunScript(cmd string, args ...string) error {
	command := []string{cmd}
	command = append(command, args...)
	err := c.RunCommand(command)
	if err != nil {
		return fmt.Errorf("could run %s: %w", command, err)
	}
	return nil
}

func (c *Container) Setup() error {
	fmt.Printf("DEBUG: preparing container for setup\n")
	err := c.AddScript("setup.sh", "/tmp/setup.sh", setupScript)
	if err != nil {
		return fmt.Errorf("could not create setup.sh: %w", err)
	}
	err = c.AddScript("start.sh", startCommand, startScript)
	if err != nil {
		return fmt.Errorf("could not create %s: %w", startCommand, err)
	}
	fmt.Printf("DEBUG: running setup script\n")
	err = c.RunScript("/tmp/setup.sh")
	if err != nil {
		return fmt.Errorf("failed to run setup.sh script: %w", err)
	}
	return nil
}

func (c *Container) RunCommand(cmd []string) error {
	args := []string{"exec", "--tty", "--user", "root", "--workdir", c.Config.WorkspaceFolder, c.Name}
	args = append(args, cmd...)
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w - %s | %s", err, err.(*exec.ExitError).Stderr, string(out))
	}
	return nil
}

func (c *Container) Start() error {
	_, err := exec.Command("docker", "start", c.Name).Output()
	if err != nil {
		return fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	return nil
}

func (c *Container) Exec(stdin io.Reader, stdout, stderr io.Writer) error {
	var args []string
	args = append(args, "exec")
	args = append(args, "--tty", "--interactive")
	args = append(args, envToDockerArgs(c.Config.RemoteEnv)...)
	args = append(args, "--user", c.Config.RemoteUser)
	args = append(args, "--workdir", c.Config.WorkspaceFolder)
	args = append(args, c.Name)
	args = append(args, startCommand)
	fmt.Printf("DEBUG: args for exec: %v\n", args)
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

func (c *Container) Stop() error {
	_, err := exec.Command("docker", "stop", c.Name).Output()
	if err != nil {
		return fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	return nil
}

func (c *Container) Rm() error {
	_, err := exec.Command("docker", "rm", "-f", c.ID).Output()
	if err != nil {
		return fmt.Errorf("%w - %s", err, err.(*exec.ExitError).Stderr)
	}
	return nil
}
