FROM ubuntu:20.04
ARG USERNAME="dev"
ARG USER_UID=1000
ARG USER_GID=$USER_UID
ARG DOTFILES_REPO=""
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
    curl 'https://installer-brl.herokuapp.com/dotfiles@v0.1.3!!' | bash && \
    addgroup -gid ${USER_GID} ${USERNAME} && \
    adduser --home /home/${USERNAME} --uid ${USER_UID} --gid ${USER_GID} --gecos "" --disabled-password ${USERNAME} && \
    usermod -aG sudo ${USERNAME} && \
    usermod -aG docker ${USERNAME} && \
    echo "${USERNAME}  ALL=(ALL) NOPASSWD:ALL" | tee /etc/sudoers.d/${USERNAME}
USER ${USERNAME}
WORKDIR /home/${USERNAME}
ADD entrypoint /bin/entrypoint
ENTRYPOINT [ "/bin/entrypoint" ]