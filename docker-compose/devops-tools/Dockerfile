# docker build --platform linux/amd64 -t docker.io/yusufs/devops-tools:2025-08-05 .
# docker run --platform linux/amd64 --entrypoint bash docker.io/yusufs/devops-tools:2025-08-05 -c "kustomize version"
FROM docker.io/library/alpine:3.22.1
LABEL maintainer="Yusuf Syaifudin <yusuf.syaifudin@gmail.com>"

ENV KUBE_VERSION=v1.33.3
ENV ARGOCD_VERSION=v3.0.12
ENV ARGOROLLOUT_VERSION=v1.8.3
ENV CRANE_VERSION=v0.20.6
ENV KOMPOSE_VERSION=v1.36.0
ENV KUSTOMIZE_VERSION=v5.7.1

# install Kubernetes Client, ArgoCD CLI, and ArgoCD Rollouts kubectl plugins
# https://github.com/argoproj/argo-cd/issues/1561
RUN apk add \
    ca-certificates bash \
    curl bash which wget coreutils tzdata \
    git openssh zip

# install kubectl, argocd, kubectl-argo-rollouts, crane, kompose, and kustomize
RUN curl --http1.1 -sSL -o /usr/local/bin/kubectl "https://dl.k8s.io/release/${KUBE_VERSION}/bin/linux/amd64/kubectl" && \
    chmod +x /usr/local/bin/kubectl 

RUN curl --http1.1 -sSL -o /usr/local/bin/argocd "https://github.com/argoproj/argo-cd/releases/download/${ARGOCD_VERSION}/argocd-linux-amd64" && \
    chmod +x /usr/local/bin/argocd && \
    curl --http1.1 -sSL -o /usr/local/bin/kubectl-argo-rollouts "https://github.com/argoproj/argo-rollouts/releases/download/${ARGOROLLOUT_VERSION}/kubectl-argo-rollouts-linux-amd64" && \
    chmod +x /usr/local/bin/kubectl-argo-rollouts

RUN curl --http1.1 -sSL -o /tmp/go-containerregistry.tar.gz "https://github.com/google/go-containerregistry/releases/download/${CRANE_VERSION}/go-containerregistry_Linux_x86_64.tar.gz" && \
    tar -zxvf /tmp/go-containerregistry.tar.gz -C /usr/local/bin/ crane && \
    chmod +x /usr/local/bin/crane && \
    rm -rf /tmp/go-containerregistry.tar.gz

RUN curl --http1.1 -sSL -o /usr/local/bin/kompose "https://github.com/kubernetes/kompose/releases/download/${KOMPOSE_VERSION}/kompose-linux-amd64" && \
    chmod +x /usr/local/bin/kompose

RUN curl --http1.1 -sSL -o /tmp/kustomize.tar.gz "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F${KUSTOMIZE_VERSION}/kustomize_${KUSTOMIZE_VERSION}_linux_amd64.tar.gz" && \ 
    tar -zxvf /tmp/kustomize.tar.gz -C /usr/local/bin kustomize && \
    chmod +x /usr/local/bin/kustomize && \
    rm -rf /tmp/kustomize.tar.gz

ENTRYPOINT ["kubectl"]
