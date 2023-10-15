build:
	docker build -t localhost:5000/bar .

push:
	docker push localhost:5000/bar

run:
	docker run -d --publish 8080:8080 localhost/bar

stop:
	docker stop $(podman ps -q)

start-cluster:
	minikube start && minikube addons enable registry


stop-cluster:
	minikube delete --all

.PHONY: build run stop start-cluster stop-cluster
