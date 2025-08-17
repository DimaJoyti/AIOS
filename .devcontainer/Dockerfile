# Development container for AIOS
FROM mcr.microsoft.com/devcontainers/go:1.23-bullseye

# Install additional system packages
RUN apt-get update && export DEBIAN_FRONTEND=noninteractive \
    && apt-get -y install --no-install-recommends \
        # Development tools
        build-essential \
        curl \
        wget \
        unzip \
        jq \
        tree \
        htop \
        # Database clients
        postgresql-client \
        redis-tools \
        # Network tools
        netcat \
        telnet \
        # AI/ML dependencies
        python3 \
        python3-pip \
        python3-venv \
        # Monitoring tools
        prometheus-node-exporter \
    && apt-get autoremove -y && apt-get clean -y && rm -rf /var/lib/apt/lists/*

# Install Go tools
RUN go install -v github.com/ramya-rao-a/go-outline@latest \
    && go install -v github.com/cweill/gotests/gotests@latest \
    && go install -v github.com/fatih/gomodifytags@latest \
    && go install -v github.com/josharian/impl@latest \
    && go install -v github.com/haya14busa/goplay/cmd/goplay@latest \
    && go install -v github.com/go-delve/delve/cmd/dlv@latest \
    && go install -v honnef.co/go/tools/cmd/staticcheck@latest \
    && go install -v golang.org/x/tools/gopls@latest \
    && go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest \
    && go install -v golang.org/x/tools/cmd/goimports@latest \
    && go install -v github.com/securecodewarrior/gosec/v2/cmd/gosec@latest \
    && go install -v golang.org/x/vuln/cmd/govulncheck@latest \
    && go install -v github.com/google/go-licenses@latest

# Install Node.js tools globally
RUN npm install -g \
        typescript \
        @types/node \
        eslint \
        prettier \
        @next/eslint-config-next \
        license-checker \
        npm-check-updates

# Install additional development tools
RUN curl -fsSL https://get.docker.com | sh \
    && curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl \
    && curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Install database migration tool
RUN curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz \
    && mv migrate /usr/local/bin/

# Install syft for SBOM generation
RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin

# Install grype for vulnerability scanning
RUN curl -sSfL https://raw.githubusercontent.com/anchore/grype/main/install.sh | sh -s -- -b /usr/local/bin

# Install k6 for performance testing
RUN gpg -k \
    && gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69 \
    && echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | tee /etc/apt/sources.list.d/k6.list \
    && apt-get update \
    && apt-get install k6

# Install AI/ML tools
RUN pip3 install --no-cache-dir \
        jupyter \
        numpy \
        pandas \
        matplotlib \
        seaborn \
        scikit-learn \
        torch \
        transformers

# Create workspace directory
WORKDIR /workspace

# Set up shell environment
RUN echo 'export PATH=$PATH:/usr/local/go/bin:/go/bin' >> /home/vscode/.bashrc \
    && echo 'export GOPATH=/go' >> /home/vscode/.bashrc \
    && echo 'export GO111MODULE=on' >> /home/vscode/.bashrc \
    && echo 'alias ll="ls -la"' >> /home/vscode/.bashrc \
    && echo 'alias k="kubectl"' >> /home/vscode/.bashrc \
    && echo 'alias dc="docker-compose"' >> /home/vscode/.bashrc

# Set up zsh environment (if using oh-my-zsh)
RUN echo 'export PATH=$PATH:/usr/local/go/bin:/go/bin' >> /home/vscode/.zshrc \
    && echo 'export GOPATH=/go' >> /home/vscode/.zshrc \
    && echo 'export GO111MODULE=on' >> /home/vscode/.zshrc \
    && echo 'alias ll="ls -la"' >> /home/vscode/.zshrc \
    && echo 'alias k="kubectl"' >> /home/vscode/.zshrc \
    && echo 'alias dc="docker-compose"' >> /home/vscode/.zshrc

# Create directories for Go modules and cache
RUN mkdir -p /go/pkg/mod && chown -R vscode:vscode /go

# Switch to vscode user
USER vscode

# Set default shell
SHELL ["/bin/bash", "-c"]
