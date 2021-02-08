FROM golang:1.15.6 as builder
WORKDIR /go/src/
RUN GO111MODULE=on go get -v github.com/bottlerocketlabs/pair/cmd/pair
RUN GO111MODULE=on go get -v github.com/bottlerocketlabs/dotfiles/cmd/dotfiles
RUN GO111MODULE=on go get -v github.com/bottlerocketlabs/remote-pbcopy/cmd/rpbcopy

FROM ubuntu:20.04
ARG USERNAME="dev"
ARG USER_UID=1000
ARG USER_GID=$USER_UID
RUN apt update && \
    apt install -y --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    docker.io \
    git \
    sudo \
    tmux \
    vim \
    wget \
    zsh && \
    rm -rf /var/lib/apt/lists/* && \
    addgroup -gid ${USER_GID} ${USERNAME} && \
    adduser --home /home/${USERNAME} --uid ${USER_UID} --gid ${USER_GID} --gecos "" --disabled-password ${USERNAME} && \
    usermod -aG sudo ${USERNAME} && \
    usermod -aG docker ${USERNAME} && \
    echo "${USERNAME}  ALL=(ALL) NOPASSWD:ALL" | tee /etc/sudoers.d/${USERNAME}
USER ${USERNAME}
WORKDIR /home/${USERNAME}
COPY --from=builder /go/bin/pair /bin
COPY --from=builder /go/bin/rpbcopy /bin
COPY --from=builder /go/bin/dotfiles /bin
# ENV DOTFILES_REPO= # FIXME
ADD entrypoint /bin/entrypoint
ENTRYPOINT [ "/bin/entrypoint" ]