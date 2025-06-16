FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    curl \
    build-essential \
    && curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y \
    && . $HOME/.cargo/env

ENTRYPOINT ["/bin/bash", "-c", "cd /app && /bin/bash"]
