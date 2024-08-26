function load_kratix_input() {

	if [[ ! -f /kratix/input/object.yaml ]]; then
		echo "Error: /kratix/input/object.yaml not found"
		exit 1
	fi
	# ...
}

function check_binaries() {
	for cmd in yq kubectl gh sed git jq; do
		if ! command -v $cmd &>/dev/null; then
			echo "$cmd could not be found"
			exit 1
		fi
	done
}
