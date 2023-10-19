#!/bin/bash

function provision_resource_check_exceed_limit() {
  maxWatchCount=$1
  currentWatchCount=$2
  if [ -z $currentWatchCount -o -z $maxWatchCount ]
  then
    return
  fi

  if [ $currentWatchCount -ge $maxWatchCount ]
  then
    make stop
    echo "----------------------------------"
    echo "----------------------------------"
    echo "Provisioning services timed out."
    echo "Please retry."
    echo "----------------------------------"
    echo "----------------------------------"
    exit 1
  fi
}

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
  maxWatchCount=25
  provision_resource_check_exceed_limit $maxWatchCount $checkCount
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

  maxWatchCount=30
  provision_resource_check_exceed_limit $maxWatchCount $checkCount
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
  maxWatchCount=30
  provision_resource_check_exceed_limit $maxWatchCount $checkCount
done

echo "----------------------------------"
echo "elasticsearch ready"
echo "----------------------------------"

kubectl apply -f kubernetes/jaeger-tracing.yml

tilt up


