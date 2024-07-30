"#!/bin/bash"

kind create cluster --name platform
kind get kubeconfig --name platform >/tmp/platform.config
export KUBECONFIG=/tmp/platform.config
kubectl wait --for=condition=Ready=true node --all

# cert-manager
kubectl apply --filename https://github.com/cert-manager/cert-manager/releases/download/v1.12.0/cert-manager.yaml
kubectl -n cert-manager wait --for=condition=Ready po --all
# kratix
kubectl apply --filename https://github.com/syntasso/kratix/releases/latest/download/install-all-in-one.yaml
kubectl -n kratix-platform-system wait --for=condition=Ready po --all
