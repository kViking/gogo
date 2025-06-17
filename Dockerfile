FROM ubuntu:22.04

RUN apt-get update && apt-get install -y build-essential curl ca-certificates

COPY ./rustup-init /app/rustup-init
RUN chmod +x /app/rustup-init
RUN /app/rustup-init -y --no-modify-path

# Add cargo to PATH for all future RUN/CMD/ENTRYPOINT
ENV PATH="/root/.cargo/bin:${PATH}"

WORKDIR /app

ENTRYPOINT ["/bin/bash"]
