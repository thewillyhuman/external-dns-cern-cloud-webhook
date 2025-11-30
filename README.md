# ExternalDNS CERN Cloud Webhook Provider

This project implements an [ExternalDNS](https://github.com/kubernetes-sigs/external-dns) Webhook Provider for the CERN Cloud infrastructure. It allows ExternalDNS to manage DNS records at CERN by interacting with OpenStack Compute (Nova) metadata.

## Overview

At CERN, DNS records for OpenStack instances can be managed via specific metadata keys on the instances. This webhook provider acts as a bridge between ExternalDNS and the CERN OpenStack infrastructure. It translates ExternalDNS endpoint changes into OpenStack metadata updates on designated ingress nodes.

Key features:
- **Ingress Node Discovery**: Automatically identifies Kubernetes Nodes serving traffic based on a configurable label, and resolves them to OpenStack Instances by matching **Node Name** to **Instance Name**.
- **Metadata Management**: Manages `landb-alias` metadata keys, adhering to CERN's specific formatting and length constraints (handling the 254-character limit by splitting across multiple keys).
- **Load Balancing**: configured to support round-robin DNS by distributing aliases across multiple nodes with the required `--load-N-` suffix format.
- **Deterministic**: Ensures stable ordering of records to minimize unnecessary metadata updates.

## Getting Started

### Prerequisites

- Go 1.24+ (for building from source)
- Docker (for containerized deployment)
- Access to a CERN OpenStack project with permissions to update instance metadata.
- Access to the Kubernetes Cluster API (via `KUBECONFIG` or In-Cluster config).

### Installation

#### From Source

```bash
make build
```

#### Docker

To pull the latest image:

```bash
docker pull ghcr.io/thewillyhuman/external-dns-cern-cloud-webhook:latest
```

To build locally:

```bash
make build-image
```

### Configuration

The webhook is configured via command-line flags or environment variables.

| Flag | Environment Variable | Default | Description |
|------|----------------------|---------|-------------|
| `--listen-address` | `LISTEN_ADDRESS` | `0.0.0.0` | Address to listen on |
| `--listen-port` | `LISTEN_PORT` | `8888` | Port to listen on |
| `--log-level` | `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `--dry-run` | `DRY_RUN` | `false` | If true, no changes will be applied to OpenStack |
| `--ingress-label` | `INGRESS_LABEL` | `node-role.kubernetes.io/ingress` | Kubernetes Label to filter ingress nodes |
| `--os-auth-url` | `OS_AUTH_URL` | - | OpenStack Auth URL |
| `--os-project-name` | `OS_PROJECT_NAME` | - | OpenStack Project Name |
| `--os-username` | `OS_USERNAME` | - | OpenStack Username |
| `--os-password` | `OS_PASSWORD` | - | OpenStack Password |
| `--os-region-name` | `OS_REGION_NAME` | - | OpenStack Region Name |

See `external-dns-cern-cloud-webhook --help` for the full list of options.

### Deployment Example

Here is a complete Kubernetes deployment example including:
1.  **RBAC**: Permissions for the Webhook to read Nodes and for ExternalDNS to read Services/Ingresses.
2.  **Webhook Deployment**: Running the CERN Cloud provider sidecar (or standalone service).
3.  **ExternalDNS Deployment**: Configured to talk to the webhook.

Save this as `deployment.yaml` and apply with `kubectl apply -f deployment.yaml`.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: external-dns
---
# RBAC for ExternalDNS (Standard)
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: external-dns
rules:
  - apiGroups: [""]
    resources: ["services", "endpoints", "pods", "nodes"]
    verbs: ["get", "watch", "list"]
  - apiGroups: ["extensions", "networking.k8s.io"]
    resources: ["ingresses"]
    verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: external-dns-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns
subjects:
  - kind: ServiceAccount
    name: external-dns
    namespace: default
---
# RBAC for CERN Webhook (Needs to read Nodes)
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: external-dns-cern-webhook
rules:
  - apiGroups: [""]
    resources: ["nodes"]
    verbs: ["get", "watch", "list"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: external-dns-cern-webhook-viewer
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: external-dns-cern-webhook
subjects:
  - kind: ServiceAccount
    name: external-dns
    namespace: default
---
# Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: external-dns
spec:
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app: external-dns
  template:
    metadata:
      labels:
        app: external-dns
    spec:
      serviceAccountName: external-dns
      containers:
        # --- ExternalDNS Container ---
        - name: external-dns
          image: registry.k8s.io/external-dns/external-dns:v0.14.0
          args:
            - --source=service
            - --source=ingress
            - --domain-filter=cern.ch # Adjust to your domain
            - --provider=webhook
            - --webhook-provider-url=http://localhost:8888
            - --registry=txt
            - --txt-owner-id=k8s-cluster-id
          ports:
            - containerPort: 7979

        # --- CERN Cloud Webhook Sidecar ---
        - name: cern-webhook
          image: ghcr.io/thewillyhuman/external-dns-cern-cloud-webhook:latest
          args:
            - --listen-port=8888
            - --ingress-label=node-role.kubernetes.io/ingress
            - --log-level=debug
          env:
            - name: OS_AUTH_URL
              value: "https://keystone.cern.ch/v3"
            - name: OS_PROJECT_NAME
              value: "my-project"
            - name: OS_USERNAME
              value: "my-user"
            - name: OS_PASSWORD
              value: "my-password" # Recommend using Secrets
            - name: OS_USER_DOMAIN_NAME
              value: "default"
            - name: OS_PROJECT_DOMAIN_ID
              value: "default"
            - name: OS_REGION_NAME
              value: "cern"
```

In this example, the webhook runs as a **sidecar** container within the ExternalDNS pod, allowing them to communicate via `localhost`.

## Contribute

Contributions are welcome! Please follow these steps:

1.  Fork the repository.
2.  Create a feature branch (`git checkout -b feature/amazing-feature`).
3.  Commit your changes (`git commit -m 'Add some amazing feature'`).
4.  Run tests (`go test ./...`).
5.  Push to the branch (`git push origin feature/amazing-feature`).
6.  Open a Pull Request.

Please ensure your code follows the existing style and conventions.

## License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.
