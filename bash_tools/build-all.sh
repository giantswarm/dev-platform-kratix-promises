#!/bin/bash -e

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/../build-config.sh"
source "$my_dir/../bash_libs/dir-tools.sh"

if [[ -x ./build-promise.sh ]]; then
  echo "Building promise..."
  ./build-promise.sh
fi

function build_container() {
  image_name="$images_registry/$images_repo/$(basename "$1")"
  echo "Building $image_name..."
  docker build --platform=linux/amd64 -f "$dir/Dockerfile" -t "$image_name" containers
}

for_dirs containers build_container
