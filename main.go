package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/bottlerocketlabs/localpod/pkg/config"
	"github.com/bottlerocketlabs/localpod/pkg/docker"
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
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	err := flags.Parse(args[1:])
	if err != nil {
		return err
	}
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("could not get current working directory: %w", err)
	}
	if !docker.HasDocker() {
		return fmt.Errorf("could not find 'docker' on PATH")
	}
	env.Set("localWorkspaceFolder", wd)
	var cfg *config.DevContainer
	dotConfig := path.Join(wd, ".devcontainer.json")
	_, err = os.Stat(dotConfig)
	if os.IsNotExist(err) {
		cfg, err = config.DevContainerFromEnv(env)
		if err != nil {
			return fmt.Errorf("could not get config from environment: %w", err)
		}
		b, err := json.MarshalIndent(&cfg, "", "\t")
		if err != nil {
			return fmt.Errorf("could not marshal defaultConfig: %w", err)
		}
		fmt.Printf("DEBUG: writing config file: %s\n", dotConfig)
		ioutil.WriteFile(dotConfig, b, 0664)
	}
	fmt.Printf("DEBUG: reading config file: %s\n", dotConfig)
	f, err := os.Open(dotConfig)
	if err != nil {
		return fmt.Errorf("could not open file %s: %w", dotConfig, err)
	}
	cfg, err = config.DevContainerFromFile(f)
	if err != nil {
		return fmt.Errorf("could not process config file %s: %w", dotConfig, err)
	}
	err = docker.BuildImage(cfg, env, stdout, stderr)
	if err != nil {
		return fmt.Errorf("could not build image from configured dockerfile: %w", err)
	}

	container, err := docker.CreateContainer(cfg.Name, env, cfg)
	if err != nil {
		return fmt.Errorf("could not create container: %w", err)
	}
	fmt.Printf("DEBUG: container: %s\n", container.Name)
	err = container.Start()
	if err != nil {
		return fmt.Errorf("could not start container: %w", err)
	}
	err = container.Setup()
	if err != nil {
		return fmt.Errorf("could not setup container: %w", err)
	}
	err = container.Exec(stdin, stdout, stderr)
	if err != nil {
		return fmt.Errorf("could not exec container: %w", err)
	}
	if container.Config.ShutdownAction == config.StopContainer {
		err := container.Stop()
		if err != nil {
			return fmt.Errorf("could not stop container: %w", err)
		}
	}
	return nil
}
