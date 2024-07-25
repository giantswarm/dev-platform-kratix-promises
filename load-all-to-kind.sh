#!/bin/bash -e

KIND_NAME="platform"

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/build-config.sh"

shopt -s globstar nullglob
for kind_name in "${KIND_NAME[@]}"; do
	echo "Uploading to kind cluster '$kind_name'"
	for d in "$my_dir"/**/Dockerfile; do
		d=$(dirname "$d")
		if [[ "$d" =~ .*/_promise_template/.* ]]; then
			continue
		fi
		img="$images_registry/${d##*/}"
		echo "* uploading $img"
		kind load docker-image "$img" --name "$kind_name"
	done
done
