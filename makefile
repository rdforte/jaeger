start:
	./runLocal.sh

stop:
	minikube delete --all

.PHONY: start stop
