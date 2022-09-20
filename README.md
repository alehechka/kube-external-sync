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

Replication of resources is triggered on the creation/modification of Namespaces or the annotated resource in question. To turn on replication of a resource, it must be annotated with at least one of the following:

| Annotation                                    | Example                          | Description                                                                                                                                      |
| --------------------------------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `kube-external-sync.io/replicate-to`          | `my-namespace, feature-[.]*`     | A CSV list of namespace names or regex patterns to replicate the resource to.                                                                    |
| `kube-external-sync.io/replicate-to-matching` | `feature-branch,branch!=default` | A [label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) used to locate labeled namespaces to replicate to. |

### Example:

The following Service will be replicated as an ExternalName Service to all Namespaces that name match the `feature-[.]*` regex or include a label of the name `feature-branch` with any value.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx
  namespace: default
  annotations:
    kube-external-sync.io/replicate-to: 'feature-[.]*'
    kube-external-sync.io/replicate-to-matching: 'feature-branch'
spec:
  type: ClusterIP
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: nginx
    tier: web
```

<details>
    <summary>Labeled Namespace Example</summary>

```yaml
kind: Namespace
apiVersion: v1
metadata:
  name: coolnewthing
  labels:
    feature-branch: feature.coolnewthing
```

</details>

### Services

If the original resource is of type Service, than the resulting ExternalName Service will look as follows:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx
  annotations:
    kube-external-sync.io/replicated-at: '2022-09-18T20:57:44-05:00'
    kube-external-sync.io/replicated-from: default/nginx
    kube-external-sync.io/replicated-from-version: '123456'
spec:
  type: ExternalName
  ports:
    - name: http
      port: 80
      protocol: TCP
      targetPort: 80
  externalName: nginx.default.svc.cluster.local
```

> Note the `externalName` variable in the spec, this is generated specifically to point this Service to the original default/nginx Service so that all requests made in the feature branch to `http://nginx` will be resolved back to the default branch instead of failing.

### Ingresses

Ingresses are a little more complex because they typically include TLS hosts and rules with hosts that tell the load balancer where to send incoming traffic. To handle this, the controller will assume that the Namespace name correlates directly to the first subdomain of the host.

Example: Provided host is `my-app.example.com` and the new Namespace is `feature-coolnewthing`. By default, the controller will replace the first subdomain with the Namespace name, resulting in a new host of `feature-coolnewthing.example.com`.

#### Example Input:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx
  namespace: default
  annotations:
    kube-external-sync.io/replicate-to: 'feature-[.]*'
    kube-external-sync.io/replicate-to-matching: 'feature-branch'
spec:
  tls:
    - hosts:
        - my-app.example.com
  rules:
    - host: my-app.example.com
      http:
        paths:
          - backend:
              service:
                name: nginx
                port:
                  number: 80
            path: /
            pathType: Prefix
```

#### Replicated to `feature-coolnewthing` Namespace.

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: nginx
  namespace: feature-coolnewthing
  annotations:
    kube-external-sync.io/replicated-at: '2022-09-18T21:44:29-05:00'
    kube-external-sync.io/replicated-from: default/nginx
    kube-external-sync.io/replicated-from-version: '123456'
spec:
  tls:
    - hosts:
        - feature-coolnewthing.example.com
  rules:
    - host: feature-coolnewthing.example.com
      http:
        paths:
          - backend:
              service:
                name: nginx
                port:
                  number: 80
            path: /
            pathType: Prefix
```

### Other Annotation Options

#### Annotations for any resource

| Annotation                                    | Example | Description                                                                                                                                                                                                                                |
| --------------------------------------------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `kube-external-sync.io/strip-labels`          | `true`  | By default, all labels will be replicated. This annotation will strip all labels from the replicated resource. A `app.kubernetes.io/managed-by` label will always be applied to the replicated resource.                                   |
| `kube-external-sync.io/strip-annotations`     | `true`  | By default, all non-`replicate-to` annotations will be replicated. This annotation will strip all annotations from the replicated resource. A few `kube-external-sync.io/*` annotations will always be applied to the replicated resource. |
| `kube-external-sync.io/keep-owner-references` | `true`  | By default, no OwnerReferences will be replicated. This annotation will replicate all OwnerReferences from the original.                                                                                                                   |

#### Annotations for Services

| Annotation                                   | Example        | Description                                                                                                                                      |
| -------------------------------------------- | -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `kube-external-sync.io/external-name-suffix` | `traefik.mesh` | The default value is `svc.cluster.local` but if some other service mesh is being used, that can be substituted here for the ExternalName suffix. |

#### Annotations for Ingresses

| Annotation                               | Example                 | Description                                                                                                                                                                                     |
| ---------------------------------------- | ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kube-external-sync.io/top-level-domain` | `*.feature.example.com` | By default, the top-level-domain is determined from the original resource. This annotation allows that to be overridden with a custom TLD that will be applied to all TLS hosts and rule hosts. |
| `kube-external-sync.io/tld-secret-name`  | `tls-cert-secret`       | If a custom TLD is supplied that requires a secret for the TLS cert, the SecretName can be supplied with this annotation.                                                                       |
