# Custom Kubernetes Controller

A Kubernetes controller that watches custom resources and processes them through a workqueue with structured logging.

## Project Structure

```
.
├── cmd/controller/    # Entry point
├── pkg/
│   ├── config/        # Kubeconfig setup
│   ├── controller/    # Controller logic
│   └── handlers/      # Event handlers
└── manifests/         # CRDs and sample resources
```

## What it does

This controller:
- Watches for changes to custom resources in the `myk8s.io/v1` API group
- Uses informers to track resource changes and maintain cache
- Processes resources through a rate-limited workqueue
- Handles resource events with multiple worker goroutines
- Provides structured logging of operations

## Prerequisites

- Go 1.16+
- Access to a Kubernetes cluster (minikube, kind, or remote)
- kubectl configured

## Setup and Running

```bash
# Install the CRD
kubectl apply -f manifests/crd.yaml

# Build the controller
go build -o controller cmd/controller/main.go

# Run the controller
./controller
```

## Creating Sample Resources

Sample CR manifests are provided in the `manifests/` directory:

```bash
# Create sample resources
kubectl apply -f manifests/sample-cr.yaml
kubectl apply -f manifests/sample-cr2.yaml

# List samples
kubectl get samples

# Delete samples
kubectl delete -f manifests/sample-cr.yaml
```

## Controller Events Flow

1. Informer detects resource changes
2. Event handlers add resource keys to the workqueue
3. Worker goroutines process items from the queue
4. Operations are logged with context information

## Future Development

This controller currently implements the core event detection and queueing patterns, but doesn't yet include reconciliation logic. Future enhancements will include:

- Reconciliation loop implementation
- Status updates for resources
- Creation and management of dependent resources
- Custom business logic based on resource specifications

## Logs

The controller uses structured logging with prefixes to indicate the source:
- `EVENT:` - Resource events detected by informers
- `QUEUE:` - Workqueue operations
- `WORKER:` - Worker processing
