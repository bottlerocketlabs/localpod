package config

import (
	"encoding/json"
	"errors"
	"io"
)

// From https://code.visualstudio.com/docs/remote/devcontainerjson-reference#_devcontainerjson-properties

type DevContainer struct {
	Name            string            `json:"name"`
	Image           string            `json:"image"`
	Dockerfile      string            `json:"dockerfile"`
	Context         string            `json:"context"`
	Build           DevContainerBuild `json:"build,omitempty"`
	AppPort         []string          `json:"appPort"`
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
	ExecCommand     []string          `json:"execCommand"`
}

type DevContainerBuild struct {
	Dockerfile string            `json:"dockerfile"`
	Context    string            `json:"context"`
	Args       map[string]string `json:"args"`
	Target     string            `json:"target"`
}

type ShutdownAction string

const (
	None          ShutdownAction = "none"
	StopContainer                = "stopContainer"
)

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
	DefaultImage           = "stuartwarren/localpod-base:latest"
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
		ExecCommand:     []string{"/bin/sh"},
	}
}

func DevContainerFromFile(r io.Reader) (*DevContainer, error) {
	dc := DefaultDevContainer()
	d := json.NewDecoder(r)
	d.DisallowUnknownFields()
	err := d.Decode(&dc)
	return &dc, err
}

func DevContainerFromEnv(env Env) (*DevContainer, error) {
	dc := DefaultDevContainer()
	if image := env.Get("LOCALPOD_IMAGE"); image != "" {
		dc.Image = image
	}
	if dotfiles := env.Get("DOTFILES_REPO"); dotfiles != "" {
		dc.ContainerEnv["DOTFILES_REPO"] = dotfiles
	}
	return &dc, nil
}
