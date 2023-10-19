#!/bin/bash

minikube start --cpus=2 --memory=4g

kubectl create -f kubernetes/observability-namespace.yml
kubectl create -f kubernetes/cert-manager.yml

kubectl create -f kubernetes/elastic-crds.yml
kubectl create -f kubernetes/elastic-operator.yml

checkCount=0

while [ -z $(kubectl get deployment cert-manager-webhook -n cert-manager -o jsonpath="{.status.readyReplicas}") ]
do
  checkCount=$(($checkCount + 1))
  echo "checking cert manager webhook -> $checkCount times"
  sleep 3
done

echo "----------------------------------"
echo "cert manager webhook ready"
echo "----------------------------------"

kubectl create -f kubernetes/jaeger-operator.yml

checkCount=0

while [ -z $(kubectl get deployment jaeger-operator -n observability -o jsonpath="{.status.readyReplicas}") ]
do
  checkCount=$(($checkCount + 1))
  echo "checking jaeger operator -> $checkCount times"
  sleep 3

  restart_run_local $checkCount
done

echo "----------------------------------"
echo "jaeger operator ready"
echo "----------------------------------"

kubectl apply -f kubernetes/elasticsearch.yml
kubectl create secret generic jaeger-secret --from-literal=ES_USERNAME=elastic --from-literal=ES_PASSWORD=$(kubectl get secret quickstart-es-elastic-user -o go-template='{{.data.elastic | base64decode}}')

es="Pending"
checkCount=0

while [ $es != "Running" ]
do
  checkCount=$(($checkCount + 1))
  echo "checking elastic search running -> $checkCount times"
  sleep 3
  es=$( kubectl get pod quickstart-es-default-0 -o jsonpath="{.status.phase}")
done

echo "----------------------------------"
echo "elasticsearch ready"
echo "----------------------------------"

kubectl apply -f kubernetes/jaeger-tracing.yml

tilt up

# HELPERS

function restart_run_local() {
  retryCount=5
  if [ -z $1 ]
  then
    return
  fi

  if [ $1 -ge $retryCount]
  then
    make stop
    exit 1
    make start
  fi
}

