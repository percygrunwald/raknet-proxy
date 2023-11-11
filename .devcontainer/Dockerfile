FROM ubuntu:jammy

# https://asdf-vm.com/guide/getting-started.html#_1-install-dependencies
ENV ASDF_DEPS="git curl"

# https://github.com/kennyp/asdf-golang
ENV GO_DEPS="coreutils curl"

# https://github.com/pyenv/pyenv/wiki#suggested-build-environment
ENV PYTHON_DEPS="build-essential libssl-dev zlib1g-dev libbz2-dev libreadline-dev libsqlite3-dev curl libncursesw5-dev xz-utils tk-dev libxml2-dev libxmlsec1-dev libffi-dev liblzma-dev"

# Other deps
ENV OTHER_DEPS="sudo netcat socat iproute2 command-not-found"

SHELL ["/bin/bash", "-c"]

RUN apt update && \
  DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt install -y \
  ${ASDF_DEPS} \
  ${GO_DEPS} \
  ${PYTHON_DEPS} \
  ${OTHER_DEPS}

COPY ../.tool-versions .
COPY ../.tool-versions ~
COPY ../.go-deps .
COPY ../script/install_deps.sh .

RUN export ASDF_DIR=~/.asdf \
  && if [ ! -d "${ASDF_DIR}/bin" ]; then \
      git clone https://github.com/asdf-vm/asdf.git "${ASDF_DIR}" --branch v0.13.1; \
    fi \
  && echo '. "$HOME/.asdf/asdf.sh"' >> ~/.bashrc \
  && echo '. "$HOME/.asdf/completions/asdf.bash"' >> ~/.bashrc \
  && git config --global --add safe.directory '*'
RUN export ASDF_DIR=~/.asdf && ./install_deps.sh
RUN rm -rf /var/lib/apt/lists/*

CMD ["bash"]
