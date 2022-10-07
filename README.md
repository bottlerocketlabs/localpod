# localpod

a bare minimum version of [gitpod.io](https://www.gitpod.io/) for local development

start up a preconfigured/shared development environment locally

configured via environment variables or config files

vscode and localpod both use `.devcontainer.json` - https://code.visualstudio.com/docs/remote/devcontainerjson-reference (not fully supported yet)

## dependencies

* docker
* a terminal

## tips

inside the container, file within `/.ssh/*` are copied into the users home directory and given the correct permissions, an ssh-agent is started and the default private key names are added.
