# Custom Kubernetes Controller

A simple Kubernetes controller that watches custom resources and processes them through a workqueue.

## What it does

This controller:
- Watches for changes to custom resources in the `myk8s.io/v1` API group
- Uses informers to track resource changes
- Processes resources through a workqueue
- Logs resource events (add/update/delete)
- Handles resource events with multiple workers

## Prerequisites

- Go 1.16+
- Access to a Kubernetes cluster (minikube, kind, or remote)
- kubectl configured
- Custom Resource Definition installed:

```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: samples.myk8s.io
spec:
  group: myk8s.io
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                size:
                  type: integer
  scope: Namespaced
  names:
    plural: samples
    singular: sample
    kind: Sample
    shortNames:
    - sam
```

## Running the Controller

```bash
# Build the controller
go build -o controller

# Run the controller
./controller
```

## Example Usage

```bash
# Create a sample resource
kubectl apply -f - <<EOF
apiVersion: myk8s.io/v1
kind: Sample
metadata:
  name: example-sample
spec:
  size: 3
EOF

# Create another sample
kubectl apply -f - <<EOF
apiVersion: myk8s.io/v1
kind: Sample
metadata:
  name: second-sample
spec:
  size: 5
EOF

# Watch the controller logs to see the events being processed
```

The controller will detect resource events and process them through the workqueue. Check the logs to see the resources being tracked.
