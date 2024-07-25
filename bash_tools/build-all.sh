#!/bin/bash -e

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/../build-config.sh"

for dir in containers/*; do
	if [[ ! -d "$dir" || ! -f "$dir/Dockerfile" ]]; then
		continue
	fi
	image_name="$images_registry/$(basename "$dir")"
	echo "Building $image_name..."
	(
		cd "$dir"
		docker build --build-context shared=../shared -t "$image_name" .
	)
done
