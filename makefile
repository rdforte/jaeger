build-bar:
	docker build -t localhost:5000/bar .

push-bar:
	docker push localhost:5000/bar

run-bar:
	docker run -d --publish 8080:8080 localhost/bar

stop-all:
	docker stop $(docker ps -q)

addons:
	minikube addons enable ingress && minikube addons enable registry

start-cluster:
	minikube start

start-registry:
	docker run --rm -it --network=host alpine ash -c "apk add socat && socat TCP-LISTEN:5000,reuseaddr,fork TCP:$(minikube ip):5000"

deploy-local:
	./deploy.sh

start-ingress:
	minikube tunnel

stop-cluster:
	minikube delete --all

.PHONY: build run stop start-cluster stop-cluster
