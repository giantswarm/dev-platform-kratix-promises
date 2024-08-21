#!/bin/bash -e

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/../build-config.sh"

if [[ -x ./build-promise.sh ]]; then
  echo "Building promise..."
  ./build-promise.sh
fi

for dir in containers/*; do
  if [[ ! -d "$dir" || ! -f "$dir/Dockerfile" ]]; then
    continue
  fi
  image_name="$images_registry/$images_repo/$(basename "$dir")"
  echo "Building $image_name..."
  docker build -f "$dir/Dockerfile" -t "$image_name" containers
done
