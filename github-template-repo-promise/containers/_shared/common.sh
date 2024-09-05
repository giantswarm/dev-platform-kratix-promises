function load_kratix_input() {

  if [[ ! -f /kratix/input/object.yaml ]]; then
    echo "Error: /kratix/input/object.yaml not found"
    exit 1
  fi

  # load data with yq from /kratix/input/object.yaml
  objName=$(yq '.metadata.name' /kratix/input/object.yaml)
  objNamespace=$(yq '.metadata.namespace' /kratix/input/object.yaml)
  if [[ "$objNamespace" == "null" ]]; then
    objNamespace="default"
  fi
  repoName=$(yq '.spec.repository.name' /kratix/input/object.yaml | tr '[:upper:]' '[:lower:]')
  repoOwner=$(yq '.spec.repository.owner' /kratix/input/object.yaml | tr '[:upper:]' '[:lower:]')
  repoDesc="$(yq '.spec.repository.description' /kratix/input/object.yaml)"
  repoSource=$(yq '.spec.repository.templateSource' /kratix/input/object.yaml)
  repoVisibility=$(yq '.spec.repository.visibility' /kratix/input/object.yaml | tr '[:upper:]' '[:lower:]')
  tokenName=$(yq '.spec.githubTokenSecretRef.name' /kratix/input/object.yaml)
  tokenNamespace=$(yq '.spec.githubTokenSecretRef.namespace' /kratix/input/object.yaml)
  if [[ "$tokenNamespace" == "null" ]]; then
    tokenNamespace="$objNamespace"
  fi
  destinationNamespace=$(yq '.spec.destinationNamespace' /kratix/input/object.yaml)
  if [[ "$destinationNamespace" == "null" ]]; then
    destinationNamespace="default"
  fi
  regInfoCmName=$(yq '.spec.registryInfoConfigMapRef.name' /kratix/input/object.yaml)
  regInfoCmNamespace=$(yq '.spec.registryInfoConfigMapRef.namespace' /kratix/input/object.yaml)
  if [[ "$regInfoCmNamespace" == "null" ]]; then
    regInfoCmNamespace="$objNamespace"
  fi
  backstageEntityOwner=$(yq '.spec.backstageCatalogEntity.owner' /kratix/input/object.yaml)
  backstageEntityLifecycle=$(yq '.spec.backstageCatalogEntity.lifecycle' /kratix/input/object.yaml)

  # explicit export after assignemnt to avoid loosing the exit code
  # of the command invoked - this breaks `-e` behavior
  export objName
  export objNamespace
  export repoName
  export repoOwner
  export repoDesc
  export repoSource
  export repoVisibility
  export tokenName
  export tokenNamespace
  export destinationNamespace
  export regInfoCmName
  export regInfoCmNamespace
  export backstageEntityOwner
  export backstageEntityLifecycle

  echo "Input loaded from /kratix/input/object.yaml"
}

function load_gh_token() {
  TOKEN_KEY=gh_token
  if [[ -z "$GH_TOKEN" ]]; then
    # load github token from secret
    echo "GitHub token not set in \$GH_TOKEN environment variable. Trying to load from the configured Secret."
    token=$(kubectl get secret "$tokenName" -n "$tokenNamespace" -o jsonpath="{.data.$TOKEN_KEY}")
    if [[ -z "$token" ]]; then
      echo "Couldn't load GitHub access token from secret \"$tokenName\" in namespace \"$tokenNamespace\" with key \"$TOKEN_KEY\""
      exit 2
    fi
    token=$(echo "$token" | base64 -d)
    echo "GitHub access token loaded from secret \"$tokenName\" in namespace \"$tokenNamespace\""
    GH_TOKEN=$token
  else
    echo "Environment variable \$GH_TOKEN already set, using existing value"
  fi

  export GH_TOKEN
}

function load_gh_repo_view() {
  # check if the repo already exists
  set +e
  repo_json=$(gh repo view "${repoOwner}/${repoName}" --json name,owner,visibility)
  res=$?
  set -e
  if [[ $res != 0 ]]; then
    return $res
  fi

  # if the command returned an output, it is formatted as a JSON object like below
  # {
  #   "name": "kratix-test-go-1",
  #   "owner": {
  #     "id": "MDEyOk9yZ2FuaXphdGlvbjc1NTYzNDA=",
  #     "login": "giantswarm"
  #   },
  #   "visibility": "PRIVATE"
  # }
  # extract and compare name, owner, visibility
  # if they are the same, the repository is OK not we have nothing to do
  # else, we bail out with an error

  loadedRepoName=$(echo "$repo_json" | jq -r '.name' | tr '[:upper:]' '[:lower:]')
  loadedRepoOwner=$(echo "$repo_json" | jq -r '.owner.login' | tr '[:upper:]' '[:lower:]')
  loadedRepoVisibility=$(echo "$repo_json" | jq -r '.visibility' | tr '[:upper:]' '[:lower:]')

  export loadedRepoName
  export loadedRepoOwner
  export loadedRepoVisibility

  return 0
}

function write_metadata_message() {
  cat >/kratix/metadata/status.yaml <<EOF
message: "$1"
EOF
}

function check_binaries() {
  for cmd in yq kubectl gh sed git jq; do
    if ! command -v $cmd &>/dev/null; then
      echo "$cmd could not be found"
      exit 1
    fi
  done
}
