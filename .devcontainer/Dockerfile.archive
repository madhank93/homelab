ARG GO_VERSION=1.23
FROM mcr.microsoft.com/devcontainers/go:${GO_VERSION}-bookworm

ARG NODE_VERSION=22.18.0
ARG PULUMI_VERSION=3.188.0
ARG YQ_VERSION=4.47.1
ARG JUST_VERSION=1.38.0
ARG HELM_VERSION=3.18.4
ARG CDK8S_VERSION=2.200.153
ARG KUBECTL_VERSION=1.33.0
ARG K9S_VERSION=0.32.0
ARG STERN_VERSION=1.30.0
ARG KUBECTX_VERSION=0.9.5
ARG ARGOCD_VERSION=2.12.0
ARG EZA_VERSION=0.20.7
ARG FZF_VERSION=0.56.3
ARG LSD_VERSION=1.1.5
ARG DUST_VERSION=1.1.1
ARG BROOT_VERSION=1.47.0
ARG FD_VERSION=10.2.0
ARG RIPGREP_VERSION=14.1.1
ARG ZOXIDE_VERSION=0.9.6
ARG NODE_VERSION=22.18.0
ARG BAT_VERSION=0.25.0
ARG GPING_VERSION=1.20.1

USER root

# Install NVM and Node.js
RUN curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash \
    && export NVM_DIR=/usr/local/share/nvm \
    && [ -s "$NVM_DIR/nvm.sh" ] && . "$NVM_DIR/nvm.sh" \
    && nvm install ${NODE_VERSION} \
    && nvm use ${NODE_VERSION} \
    && nvm alias default ${NODE_VERSION}

# Setup NVM environment
ENV NVM_DIR="/usr/local/share/nvm"
ENV PATH="/usr/local/share/nvm/versions/node/v${NODE_VERSION}/bin:${PATH}"


# Essential packages (cleaned up)
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
    curl ca-certificates jq unzip git make bash-completion \
    python3 python3-venv python3-pip glances \
    openssh-client rsync iputils-ping \
    build-essential pkg-config libssl-dev \
    && rm -rf /var/lib/apt/lists/*

# Install Bat
RUN curl -fsSL "https://github.com/sharkdp/bat/releases/download/v${BAT_VERSION}/bat-v${BAT_VERSION}-aarch64-unknown-linux-gnu.tar.gz" \
    -o /tmp/bat.tar.gz \
    && tar -xzf /tmp/bat.tar.gz --strip-components=1 -C /usr/local/bin \
    && chmod +x /usr/local/bin/bat \
    && rm /tmp/bat.tar.gz

# Install EZA
RUN curl -fsSL "https://github.com/eza-community/eza/releases/download/v${EZA_VERSION}/eza_aarch64-unknown-linux-gnu.tar.gz" \
    -o /tmp/eza.tar.gz \
    && tar -tf /tmp/eza.tar.gz \
    && tar -xzf /tmp/eza.tar.gz -C /usr/local/bin --strip-components=0 \
    && chmod +x /usr/local/bin/eza \
    && rm /tmp/eza.tar.gz

# Install FZF
RUN curl -fsSL "https://github.com/junegunn/fzf/releases/download/v${FZF_VERSION}/fzf-${FZF_VERSION}-linux_amd64.tar.gz" \
    | tar -xz -C /usr/local/bin fzf

# Install LSD
RUN curl -fsSL "https://github.com/lsd-rs/lsd/releases/download/v${LSD_VERSION}/lsd-v${LSD_VERSION}-aarch64-unknown-linux-gnu.tar.gz" \
    | tar -xz --strip-components=1 -C /usr/local/bin lsd-v${LSD_VERSION}-aarch64-unknown-linux-gnu/lsd

# Install Dust    
RUN curl -fsSL "https://github.com/bootandy/dust/releases/download/v${DUST_VERSION}/dust-v${DUST_VERSION}-aarch64-unknown-linux-gnu.tar.gz" \
    | tar -xz --strip-components=1 -C /usr/local/bin dust-v${DUST_VERSION}-aarch64-unknown-linux-gnu/dust

# Install Broot
RUN curl -fsSL "https://github.com/Canop/broot/releases/download/v${BROOT_VERSION}/broot_${BROOT_VERSION}.zip" \
    -o /tmp/broot.zip \
    && unzip /tmp/broot.zip -d /tmp/broot_extract \
    && find /tmp/broot_extract -name "broot" -type f -executable -exec cp {} /usr/local/bin/ \; \
    && chmod +x /usr/local/bin/broot \
    && rm -rf /tmp/broot.zip /tmp/broot_extract

# Install fd
RUN curl -fsSL "https://github.com/sharkdp/fd/releases/download/v${FD_VERSION}/fd-v${FD_VERSION}-aarch64-unknown-linux-gnu.tar.gz" \
    | tar -xz --strip-components=1 -C /usr/local/bin fd-v${FD_VERSION}-aarch64-unknown-linux-gnu/fd

# Install Ripgrep
RUN curl -fsSL "https://github.com/BurntSushi/ripgrep/releases/download/${RIPGREP_VERSION}/ripgrep-${RIPGREP_VERSION}-x86_64-unknown-linux-musl.tar.gz" \
    | tar -xz --strip-components=1 -C /usr/local/bin ripgrep-${RIPGREP_VERSION}-x86_64-unknown-linux-musl/rg

# Install Zoxide
RUN curl -fsSL "https://github.com/ajeetdsouza/zoxide/releases/download/v${ZOXIDE_VERSION}/zoxide-${ZOXIDE_VERSION}-x86_64-unknown-linux-musl.tar.gz" \
    | tar -xz -C /usr/local/bin zoxide

# Install gtop
RUN npm install -g gtop

# Install gping binary directly
RUN curl -fsSL "https://github.com/orf/gping/releases/latest/download/gping-Linux-gnu-arm64.tar.gz" \
    | tar -xz -C /usr/local/bin \
    && chmod +x /usr/local/bin/gping

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

# Install k9s (Kubernetes CLI UI)
RUN curl -fsSL "https://github.com/derailed/k9s/releases/download/v${K9S_VERSION}/k9s_Linux_amd64.tar.gz" \
    | tar -xz -C /usr/local/bin k9s

# Install stern (multi-pod log tailing)
RUN curl -fsSL "https://github.com/stern/stern/releases/download/v${STERN_VERSION}/stern_${STERN_VERSION}_linux_amd64.tar.gz" \
    | tar -xz -C /usr/local/bin stern \
    && chmod +x /usr/local/bin/stern

# Install kubectx + kubens (context switching)
RUN curl -fsSL "https://github.com/ahmetb/kubectx/releases/download/v${KUBECTX_VERSION}/kubectx_v${KUBECTX_VERSION}_linux_x86_64.tar.gz" \
    | tar -xz -C /usr/local/bin kubectx

# Install ArgoCD CLI
RUN curl -fsSL -o /usr/local/bin/argocd \
    "https://github.com/argoproj/argo-cd/releases/download/v${ARGOCD_VERSION}/argocd-linux-amd64" \
    && chmod +x /usr/local/bin/argocd

# Setup auto completions
RUN just --completions bash > /etc/bash_completion.d/just \
    && echo 'complete -C /usr/local/bin/pulumi pulumi' > /etc/bash_completion.d/pulumi \
    && kubectl completion bash > /etc/bash_completion.d/kubectl \
    && helm completion bash > /etc/bash_completion.d/helm \
    && argocd completion bash > /etc/bash_completion.d/argocd

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
RUN go install github.com/go-delve/delve/cmd/dlv@v1.23.0 \
    && go install mvdan.cc/gofumpt@v0.6.0 \
    && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.3


# Initialize zoxide for bash
RUN echo 'eval "$(zoxide init bash)"' >> /home/vscode/.bashrc
