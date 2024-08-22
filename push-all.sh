#!/usr/bin/env bash

docker push gsoci.azurecr.io/giantswarm/github-cli-clone-template-repo-pipeline:latest
docker push gsoci.azurecr.io/giantswarm/github-cli-template-values-pipeline:latest

docker push gsoci.azurecr.io/giantswarm/appdeployment-template-pipeline:latest
docker push gsoci.azurecr.io/giantswarm/check-if-infra-ready-pipeline:latest
docker push gsoci.azurecr.io/giantswarm/provision-infrastructure-pipeline:latest
