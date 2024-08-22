#!/bin/bash

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/../bash_libs/dir-tools.sh"

if [[ ! -f promise.yaml ]]; then
  echo "promise.yaml not found"
  exit 1
fi

if [[ ! -x kubeconform ]]; then
  echo "Kubeconform not found. Please install from https://github.com/yannh/kubeconform"
  exit 1
fi

if [[ ! -x python ]]; then
  echo "Python not found. Please install from https://www.python.org/"
  exit 1
fi

if [[ ! -d tmp/ ]]; then
  mkdir tmp/
fi

yq '.spec.api' promise.yaml >tmp/api.yaml
python ../openapi2jsonschema.py tmp/api.yaml
mv schema.json tmp/

kubeconform -schema-location tmp/schema.json -summary resource-request.yaml

function validate() {

}

function for_containers() {
  for_dirs . validate
}

for_dirs containers for_containers
