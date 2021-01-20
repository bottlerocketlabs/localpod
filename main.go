package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/stuart-warren/localpod/pkg/config"
	"github.com/stuart-warren/localpod/pkg/docker"
)

// main
func main() {
	err := Run(os.Args, config.NewEnv(os.Environ()), os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		log.Fatalf("ERROR: %s", err)
	}
}

// Run is the main thread but separated out so easier to test
func Run(args []string, env config.Env, stdin io.Reader, stdout, stderr io.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}
	env.Set("localWorkspaceFolder", wd)
	var (
		cfg           *config.DevContainer
		containerName = "localpod1"
	)
	dotConfig := path.Join(wd, ".devcontainer.json")
	_, err = os.Stat(dotConfig)
	if !os.IsNotExist(err) {
		f, err := os.Open(dotConfig)
		if err != nil {
			return fmt.Errorf("could not open file %s: %w", dotConfig, err)
		}
		cfg, err = config.DevContainerFromFile(f)
		if err != nil {
			return fmt.Errorf("could not process config file %s: %w", dotConfig, err)
		}
	}
	if cfg == nil {
		cfg, err = config.DevContainerFromEnv(env)
		if err != nil {
			return fmt.Errorf("could not get config from environment: %w", err)
		}
	}
	if !docker.HasDocker() {
		return fmt.Errorf("could not find 'docker' on PATH")
	}
	// TODO: check if container already exists
	container, err := docker.CreateContainer(containerName, env, cfg)
	if err != nil {
		return fmt.Errorf("could not create container: %w", err)
	}
	fmt.Printf("DEBUG: created container: %s", container.Name)
	err = container.Start()
	if err != nil {
		return fmt.Errorf("could not start container: %w", err)
	}
	err = container.Exec(stdin, stdout, stderr)
	if err != nil {
		return fmt.Errorf("could not exec container: %w", err)
	}
	// TODO: run shutdownAction
	return nil
}
