# POC for Jaeger
## Getting Started
1. install [Docker Desktop](https://www.docker.com/products/docker-desktop/). Used for building images
2. install [Minikube](https://minikube.sigs.k8s.io/docs/start/). Used for running local K8s cluster
3. install [Tilt](https://tilt.dev). Used for hot reloading foo and bar services.

## Running Project
In the project root directory run command:
```text
make start
```
To delete the cluster and all its resources run:
```text
make stop
```
