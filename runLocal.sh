#!/bin/bash

minikube start --cpus=2 --memory=4g

kubectl create -f kubernetes/observability-namespace.yml
kubectl create -f kubernetes/cert-manager.yml

kubectl create -f kubernetes/elastic-crds.yml
kubectl create -f kubernetes/elastic-operator.yml

webhookReady=false
checkCount=0

while [ $webhookReady == false ]
do
  checkCount=$(($checkCount + 1))
  echo "checking cert manager webhook -> $checkCount times"
  sleep 3
  wh=$(kubectl get deployment cert-manager-webhook -n cert-manager -o jsonpath="{.status.readyReplicas}")

  if [ ! -z $wh ]
  then
    webhookReady=true
  fi
done

echo "----------------------------------"
echo "cert manager webhook ready"
echo "----------------------------------"

kubectl create -f kubernetes/jaeger-operator.yml

operator=false
checkCount=0

while [ $operator == false ]
do
  checkCount=$(($checkCount + 1))
  echo "checking jaeger operator -> $checkCount times"
  sleep 3
  wh=$(kubectl get deployment jaeger-operator -n observability -o jsonpath="{.status.readyReplicas}")

  if [ ! -z $wh ]
  then
    operator=true
  fi
done

echo "----------------------------------"
echo "jaeger operator ready"
echo "----------------------------------"

kubectl apply -f kubernetes/elasticsearch.yml
kubectl create secret generic jaeger-secret --from-literal=ES_USERNAME=elastic --from-literal=ES_PASSWORD=$(kubectl get secret quickstart-es-elastic-user -o go-template='{{.data.elastic | base64decode}}')

#kubectl apply -f kubernetes/jaeger-tracing.yml

tilt up

