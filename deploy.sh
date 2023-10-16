#!/bin/bash

docker build -t localhost:5000/bar .
docker push localhost:5000/bar

kubectl create -f kubernetes/bar-deployment.yml
kubectl create -f kubernetes/bar-service.yml
kubectl create -f kubernetes/app-ingress.yml

minikube tunnel
