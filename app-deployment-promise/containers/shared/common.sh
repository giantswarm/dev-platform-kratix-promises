export KRATIX_META="/kratix/metadata/status.yaml"
export KRATIX_INPUT="/kratix/input/object.yaml"

function write_metadata_message() {
  echo "Setting status: '$1'"
  cat >/kratix/metadata/status.yaml <<EOF
message: "$1"
EOF
}

function check_binaries() {
  for cmd in "$@"; do
    if ! command -v "$cmd" &>/dev/null; then
      echo "$cmd could not be found"
      exit 1
    fi
  done
}

# copies values from a yq path in one file to yq path in another file
# args: [source_path] [destination_path] [source_file] [destination_file]
function yq_copy_if_exists() {
  local source_path=$1
  local destination_path=$2
  local source_file=$3
  local destination_file=$4
  local value
  value=$(yq "$source_path" "$source_file")
  if [[ "$value" != "null" ]]; then
    export VAR=$value
    yq -i "$destination_path" "$destination_file"
  fi
}
