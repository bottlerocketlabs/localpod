FROM golang:1.15.6 as builder
WORKDIR /go/src/
RUN go build -ldflags="-s -w -X main.version=$(git tag --points-at HEAD) -X main.commit=$(git rev-parse --short HEAD)" ./cmd/pair
RUN go get -v github.com/stuart-warren/remote-pbcopy/cmd/pbcopy

FROM ubuntu:20.04
ENV UNAME="pair"
RUN apt update && \
    apt install -y \
    bash \
    ca-certificates \
    curl \
    git \
    sudo \
    tmux \
    vim \
    wget \
    zsh && \
    adduser --home /home/$UNAME --gecos "" --disabled-password $UNAME && \
    usermod -aG sudo $UNAME && \
    echo "$UNAME  ALL=(ALL) NOPASSWD:ALL" | tee /etc/sudoers.d/$UNAME
USER $UNAME
WORKDIR /home/$UNAME
COPY --from=builder /go/bin/pair /bin
COPY --from=builder /go/bin/pbcopy /bin
# ENV DOTFILES_REPO= # FIXME
ADD entrypoint /bin/entrypoint
ENTRYPOINT [ "/bin/entrypoint" ]