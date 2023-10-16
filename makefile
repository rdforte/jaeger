start:
	minikube start && tilt up

stop:
	minikube delete --all

.PHONY: start stop
