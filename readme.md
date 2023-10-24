# POC for Jaeger
## Getting Started
1. Install [Docker Desktop](https://www.docker.com/products/docker-desktop/). Used for building images
2. Install [Minikube](https://minikube.sigs.k8s.io/docs/start/). Used for running local K8s cluster
3. Install [Tilt](https://tilt.dev). Used for hot reloading foo and bar services.
4. Make sure you have `Make` installed on your OS. Will need this to run start/stop scripts.

## Running Project
In the project root directory run command:
```text
make start
```
To delete the cluster and all its resources run:
```text
make stop
```

To monitor **foo** and **bar** services click `Spacebar` in terminal or open [Tilt Desktop](http://localhost:10350/).
You will be able to see traces exported by OpenTelemetry as well as some custom logs from foo and bar services.

To send a message to the **bar service**:
```text
curl localhost:8080/bar
```
This will send a message to the bar service which will then send a message to the foo service.

## Accessing Jaeger UI
enable minikube ingress:
```text
minikube addons enable ingress
```
create network tunnel:
```text
minikube tunnel
```
### open [Jaeger UI](http://localhost/search)

## Common Issues
#### Jaeger Operator Timeout
Deploying the jaeger operator takes a bit of time to deploy, and sometimes it times out trying to find the volume to mount. If after
30 attempts of checking the Jaeger Operator it is still initialising the container then kubernetes cluster will cancel deploying the
remainder of the resources and stop the cluster. You will need to rerun `make start`.


#### ElasticSearch (ES) v8 Support
Jaeger currently does not support v8 of ElasticSearch, so we have to use v7 of ES.
[Github ElasticSearch 8.x support PR](https://github.com/jaegertracing/jaeger/issues/3571)



