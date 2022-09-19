# Feature branch external name creation for Kubernetes

This repository contains a custom Kubernetes client-go controller that can be used to make Services and Ingresses available in feature branch namespaces with ExternalName Services and customized wildcard top-level domains.

## Deployment

Using Helm

```shell
helm install kube-external-sync https://github.com/alehechka/kube-external-sync/releases/download/v1.0.0/kube-external-sync-1.0.0.tgz --namespace kube-external-sync --create-namespace
```

Using kubectl

```bash
# create the namespace
kubectl create namespace kube-external-sync

# deploy the application's resource
kubectl apply -f https://github.com/alehechka/kube-external-sync/releases/download/v1.0.0/kube-external-sync.yaml
```

## Usage
