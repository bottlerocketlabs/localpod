# localpod

a bare minimum version of [gitpod.io](https://www.gitpod.io/) for local development

start up a preconfigured/shared development environment locally

configured via environment variables, config files and specially named dockerfiles

if a `Dockerfile.localpod` exists in the current directory localpod will try to build and use it

if a `~/.localpod.yaml` file exists localpod will read and use it for defaults

vscode uses `.devcontainer.json` - https://code.visualstudio.com/docs/remote/devcontainerjson-reference 

if environment variables are set, they will set/override configuration from above
```sh
LOCALPOD_IMAGE      # docker image to pull and use as base environment
DOTFILES_REPO       # git repository to clone and use (run /install) to configure user environment
LOCALPOD_COPY       # list of local paths to copy/mount into the running container
LOCALPOD_ENTRYPOINT # default shell/command to run on docker exec
```

## dependencies

* docker
* a terminal
