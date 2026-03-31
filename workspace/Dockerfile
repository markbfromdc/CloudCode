# Workspace Base Image for Cloud IDE
# This Dockerfile builds the base image that every user workspace container runs.
# It includes Node.js, Python, and Go development environments, along with
# common developer tools. The image is optimized for size while maintaining
# a complete development experience.

FROM ubuntu:22.04

LABEL maintainer="cloudcode-team"
LABEL description="Premium Cloud IDE workspace base image"

# Prevent interactive prompts during package installation.
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies and common developer tools.
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Core utilities
    bash \
    curl \
    wget \
    git \
    ca-certificates \
    gnupg \
    lsb-release \
    sudo \
    # Build tools
    build-essential \
    pkg-config \
    cmake \
    # Editor and terminal tools
    vim \
    nano \
    tmux \
    htop \
    # Network tools
    net-tools \
    iputils-ping \
    dnsutils \
    openssh-client \
    # Archive tools
    zip \
    unzip \
    tar \
    # Database clients
    postgresql-client \
    redis-tools \
    # Python build deps
    libssl-dev \
    zlib1g-dev \
    libbz2-dev \
    libreadline-dev \
    libsqlite3-dev \
    libffi-dev \
    liblzma-dev \
    && rm -rf /var/lib/apt/lists/*

# ---- Node.js 20 LTS ----
RUN curl -fsSL https://deb.nodesource.com/setup_20.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g npm@latest \
    && npm install -g yarn pnpm typescript ts-node \
    && rm -rf /var/lib/apt/lists/*

# ---- Python 3.11 ----
RUN apt-get update && apt-get install -y --no-install-recommends \
    python3.11 \
    python3.11-venv \
    python3.11-dev \
    python3-pip \
    && update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.11 1 \
    && update-alternatives --install /usr/bin/python python /usr/bin/python3.11 1 \
    && pip3 install --no-cache-dir --upgrade pip setuptools wheel \
    && rm -rf /var/lib/apt/lists/*

# ---- Go 1.22 ----
ENV GO_VERSION=1.22.1
RUN curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:/home/developer/go/bin:${PATH}"
ENV GOPATH="/home/developer/go"

# ---- Docker CLI (for Docker-in-Docker support) ----
RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg \
    && echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
    | tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && apt-get update \
    && apt-get install -y --no-install-recommends docker-ce-cli \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root developer user with sudo access.
RUN useradd -m -s /bin/bash -G sudo developer \
    && echo "developer ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

# Set up workspace directory.
RUN mkdir -p /workspace && chown developer:developer /workspace

# Switch to non-root user.
USER developer
WORKDIR /workspace

# Install common Go tools.
RUN go install golang.org/x/tools/gopls@latest \
    && go install github.com/go-delve/delve/cmd/dlv@latest \
    && go install golang.org/x/lint/golint@latest

# Configure shell.
RUN echo 'export PS1="\[\033[01;32m\]\u@cloudide\[\033[00m\]:\[\033[01;34m\]\w\[\033[00m\]\$ "' >> ~/.bashrc

# Health check: ensure the shell is responsive.
HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
    CMD ["bash", "-c", "echo ok"]

# Default command: start an interactive bash shell.
CMD ["/bin/bash"]
