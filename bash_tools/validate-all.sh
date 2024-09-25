#!/bin/bash -e

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/../bash_libs/dir-tools.sh"

if [[ ! -f promise.yaml ]]; then
  echo "promise.yaml not found"
  exit 1
fi

if ! command -v kubeconform &>/dev/null; then
  echo "Kubeconform not found. Please install from https://github.com/yannh/kubeconform"
  exit 1
fi

if ! command -v python &>/dev/null; then
  echo "Python not found. Please install from https://www.python.org/"
  exit 1
fi

if [[ ! -d tmp/ ]]; then
  mkdir tmp/
fi

yq '.spec.api' promise.yaml >tmp/api.yaml
python ../openapi2jsonschema.py tmp/api.yaml
mv schema.json tmp/

echo "Sample resource validation: resource-request.yaml"
kubeconform -schema-location tmp/schema.json resource-request.yaml

function validate() {
  echo "Test input validation: $1/input/object.yaml"
  kubeconform -schema-location tmp/schema.json "$1/input/object.yaml"
}

function for_containers() {
  for_dirs "$1/tests" validate
}

function copy_resource_request_example() {
  if [[ $# != 1 ]]; then
    echo "Usage: $0 [dir]"
    exit 1
  fi
  dir="$1"
  if [[ ! -f "$dir/Dockerfile" || ! -d "$dir/tests/resource_request_example" ]]; then
    return
  fi
  src="resource-request.yaml"
  dst="$dir/tests/resource_request_example/input/object.yaml"
  echo "Copying $src to $dst"
  cp "$src" "$dst"
}

for_dirs "containers" copy_resource_request_example
for_dirs containers for_containers
