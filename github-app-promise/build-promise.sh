#!/bin/bash

# build the promise
yq -i '.spec.api.spec.versions.[0].schema.openAPIV3Schema.properties.spec.properties.appDeployment = (
  load("../app-deployment-promise/promise.yaml").spec.api.spec.versions.[] 
  | select (.name == "v1beta1") 
  | .schema.openAPIV3Schema.properties.spec 
  | del(.properties.statusConfigMapReference) 
  | del(.required.[] | select(. == "statusConfigMapReference"))
)' promise.yaml
yq -i '.spec.api.spec.versions.[0].schema.openAPIV3Schema.properties.spec.properties.githubRepo = (
  load("../github-template-repo-promise/promise.yaml").spec.api.spec.versions.[] 
  | select (.name == "v1beta1") 
  | .schema.openAPIV3Schema.properties.spec
)' promise.yaml

# build the example
yq -i '.spec.githubRepo.spec = (
  load("../github-template-repo-promise/resource-request.yaml").spec
)' resource-request.yaml
yq -i '.spec.appDeployment.spec = (
  load("../app-deployment-promise/resource-request.yaml").spec
  | del(.statusConfigMapReference) 
)' resource-request.yaml
