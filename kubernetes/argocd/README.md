# ArgoCD

## Installation

```shell
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/v2.7.5/manifests/install.yaml
```

Do `kubectl port-forward svc/argocd-server -n argocd 8080:443` then run ArgoCD CLI to get the initial password: `argocd admin initial-password -n argocd`.

Use this password for Dashboard login with username `admin`.