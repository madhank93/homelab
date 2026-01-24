ARG GO_VERSION=1.25
FROM mcr.microsoft.com/devcontainers/go:${GO_VERSION}-bookworm

ARG NODE_VERSION=24.13.0
ARG NVM_VERSION=0.40.3
ARG PULUMI_VERSION=3.216.0
ARG YQ_VERSION=4.50.1
ARG JUST_VERSION=1.46.0
ARG HELM_VERSION=3.20.0
ARG CDK8S_VERSION=2.203.5
ARG KUBECTL_VERSION=1.35.0
ARG DELVE_VERSION=1.26.0
ARG GOFUMPT_VERSION=0.9.2
ARG GOLANGCI_LINT_VERSION=2.8.0

USER root

# Install NVM and Node.js
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v${NVM_VERSION}/install.sh | bash \
    && export NVM_DIR=/usr/local/share/nvm \
    && [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh" \
    && nvm install ${NODE_VERSION} \
    && nvm use ${NODE_VERSION} \
    && nvm alias default ${NODE_VERSION}

# Setup NVM environment
ENV NVM_DIR="/usr/local/share/nvm"
ENV PATH="/usr/local/share/nvm/versions/node/v${NODE_VERSION}/bin:${PATH}"

# Essential packages (cleaned up)
RUN rm -f /etc/apt/sources.list.d/yarn.list \
    && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    curl ca-certificates jq unzip git make bash-completion \
    python3 python3-venv python3-pip \
    openssh-client rsync iputils-ping \
    build-essential pkg-config libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Helm
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod 700 get_helm.sh \
    && ./get_helm.sh --version v${HELM_VERSION} \
    && rm get_helm.sh

# Install YQ
RUN curl -fsSL -o /usr/local/bin/yq \
    "https://github.com/mikefarah/yq/releases/download/v${YQ_VERSION}/yq_linux_amd64" \
    && chmod +x /usr/local/bin/yq

# Install Just
RUN curl -fsSL \
    "https://github.com/casey/just/releases/download/${JUST_VERSION}/just-${JUST_VERSION}-x86_64-unknown-linux-musl.tar.gz" \
    | tar -xz -C /usr/local/bin just

# Install Pulumi
RUN curl -fsSL \
    "https://get.pulumi.com/releases/sdk/pulumi-v${PULUMI_VERSION}-linux-x64.tar.gz" \
    -o /tmp/pulumi.tgz \
    && tar -xzf /tmp/pulumi.tgz -C /usr/local \
    && ln -sf /usr/local/pulumi/pulumi /usr/local/bin/pulumi \
    && ln -sf /usr/local/pulumi/* /usr/local/bin/ \
    && rm -f /tmp/pulumi.tgz

# Install cdk8s CLI
RUN npm install -g cdk8s-cli@${CDK8S_VERSION}

# Install kubectl
RUN curl -fsSL -o /usr/local/bin/kubectl \
    "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl" \
    && chmod +x /usr/local/bin/kubectl

# Setup auto completions
RUN just --completions bash > /etc/bash_completion.d/just \
    && echo 'complete -C /usr/local/bin/pulumi pulumi' > /etc/bash_completion.d/pulumi \
    && kubectl completion bash > /etc/bash_completion.d/kubectl \
    && helm completion bash > /etc/bash_completion.d/helm

# Fix Go module cache permissions
RUN mkdir -p /home/vscode/go/pkg/mod/cache /home/vscode/go/pkg/mod/download \
    && chown -R vscode:vscode /home/vscode/go/pkg/mod \
    && chmod -R u+rwX /home/vscode/go/pkg/mod

# Prepare SSH directory (as root)
RUN mkdir -p /home/vscode/.ssh && chmod 700 /home/vscode/.ssh \
    && chown -R vscode:vscode /home/vscode/.ssh

# Switch to dev user
USER vscode

# ENV
ENV PATH="/home/vscode/go/bin:/usr/local/bin:/usr/local/share/nvm/versions/node/v${NODE_VERSION}/bin:${PATH}"

# Install Go tools (as vscode user)

RUN go install github.com/go-delve/delve/cmd/dlv@v${DELVE_VERSION} \
    && go install mvdan.cc/gofumpt@v${GOFUMPT_VERSION} \
    && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v${GOLANGCI_LINT_VERSION}