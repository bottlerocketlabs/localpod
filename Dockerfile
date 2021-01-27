FROM golang:1.15.6 as builder
WORKDIR /go/src/
RUN GO111MODULE=on go get -v github.com/bottlerocketlabs/pair/cmd/pair
RUN GO111MODULE=on go get -v github.com/bottlerocketlabs/dotfiles/cmd/dotfiles
RUN GO111MODULE=on go get -v github.com/bottlerocketlabs/remote-pbcopy/cmd/pbcopy

FROM ubuntu:20.04
ENV UNAME="dev"
RUN apt update && \
    apt install -y --no-install-recommends \
    bash \
    ca-certificates \
    curl \
    git \
    sudo \
    tmux \
    vim \
    wget \
    zsh && \
    rm -rf /var/lib/apt/lists/* && \
    adduser --home /home/$UNAME --gecos "" --disabled-password $UNAME && \
    usermod -aG sudo $UNAME && \
    echo "$UNAME  ALL=(ALL) NOPASSWD:ALL" | tee /etc/sudoers.d/$UNAME
USER $UNAME
WORKDIR /home/$UNAME
COPY --from=builder /go/bin/pair /bin
COPY --from=builder /go/bin/pbcopy /bin
COPY --from=builder /go/bin/dotfiles /bin
# ENV DOTFILES_REPO= # FIXME
ADD entrypoint /bin/entrypoint
ENTRYPOINT [ "/bin/entrypoint" ]