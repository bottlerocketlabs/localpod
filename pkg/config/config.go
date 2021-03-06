package config

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// From https://code.visualstudio.com/docs/remote/devcontainerjson-reference#_devcontainerjson-properties

type DevContainer struct {
	Name  string `json:"name"`
	Image string `json:"image"`
	// Dockerfile      string            `json:"dockerfile"`
	// Context         string            `json:"context"`
	Build DevContainerBuild `json:"build,omitempty"`
	// AppPort         []string          `json:"appPort"`
	ContainerEnv    map[string]string `json:"containerEnv"`
	RemoteEnv       map[string]string `json:"remoteEnv"`
	ContainerUser   string            `json:"containerUser"`
	RemoteUser      string            `json:"remoteUser"`
	Mounts          []string          `json:"mounts"`
	WorkspaceMount  string            `json:"workspaceMount"`
	WorkspaceFolder string            `json:"workspaceFolder"`
	RunArgs         []string          `json:"runArgs"`
	OverrideCommand bool              `json:"overrideCommand"`
	ShutdownAction  ShutdownAction    `json:"shutdownAction"`
}

type DevContainerBuild struct {
	Dockerfile string            `json:"dockerfile,omitempty"`
	Context    string            `json:"context,omitempty"`
	Args       map[string]string `json:"args,omitempty"`
	Target     string            `json:"target,omitempty"`
}

type ShutdownAction string

const (
	None          ShutdownAction = "none"
	StopContainer                = "stopContainer"
)

func BuildEnv(env map[string]string) []string {
	out := []string{}
	for k, v := range env {
		out = append(out, fmt.Sprintf("%s=%q", k, v))
	}
	return out
}

func (sa *ShutdownAction) UnmarshalJSON(b []byte) error {
	var s string
	json.Unmarshal(b, &s)
	shutdownAction := ShutdownAction(s)
	switch shutdownAction {
	case None, StopContainer:
		*sa = shutdownAction
		return nil
	}
	return errors.New("Invalid ShutdownAction option")
}

const (
	DefaultName            = "localpod"
	DefaultImage           = "docker.io/bottlerocketlabs/localpod-base:latest"
	DefaultWorkspaceMount  = "source=${localWorkspaceFolder},target=/workspace,type=bind,consistency=cached"
	DefaultWorkspaceFolder = "/workspace"
	DefaultRemoteUser      = "dev"
	DefaultContainerUser   = "root"
)

func DefaultDevContainer() DevContainer {
	return DevContainer{
		Name:            DefaultName,
		Image:           DefaultImage,
		ContainerEnv:    map[string]string{},
		RemoteEnv:       map[string]string{},
		ContainerUser:   DefaultContainerUser,
		RemoteUser:      DefaultRemoteUser,
		Mounts:          []string{},
		WorkspaceMount:  DefaultWorkspaceMount,
		WorkspaceFolder: DefaultWorkspaceFolder,
		RunArgs:         []string{},
		OverrideCommand: true,
		ShutdownAction:  StopContainer,
	}
}

func DevContainerFromFile(r io.Reader) (*DevContainer, error) {
	dc := DefaultDevContainer()
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	err := d.Decode(&dc)
	return &dc, err
}

func (dc *DevContainer) AddConfigFromEnv(env Env) error {
	if image := env.Get("LOCALPOD_IMAGE"); image != "" {
		dc.Image = image
	}
	if dotfiles := env.Get("DOTFILES_REPO"); dotfiles != "" {
		dc.ContainerEnv["DOTFILES_REPO"] = dotfiles
	}
	if mounts := env.Get("LOCALPOD_MOUNTS"); mounts != "" {
		dc.Mounts = append(dc.Mounts, strings.Split(mounts, ";")...)
	}
	if envVars := env.Get("LOCALPOD_ENV_VARS"); envVars != "" {
		ev := NewEnv(strings.Split(envVars, ";"))
		for k, v := range ev.m {
			dc.RemoteEnv[k] = v
		}
	}
	return nil
}

func DevContainerFromEnv(env Env) (*DevContainer, error) {
	dc := DefaultDevContainer()
	dc.AddConfigFromEnv(env)
	return &dc, nil
}

// SHA1 returns the hash with base64 encoding for the configuration
func (cfg *DevContainer) SHA1() string {
	b, _ := json.Marshal(&cfg)
	configHash := sha1.Sum(b)
	return base64.URLEncoding.EncodeToString(configHash[:])
}
