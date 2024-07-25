#!/bin/bash -e

my_dir=$(dirname -- "$(readlink -f -- "$0")")

"$my_dir/build-all.sh"
"$my_dir/load-all-to-kind.sh"
for d in "$my_dir"/**/promise.yaml; do
	if [[ "$d" =~ .*/_promise_template/promise.yaml ]]; then
		continue
	fi
	kubectl apply -f "$d"
done
