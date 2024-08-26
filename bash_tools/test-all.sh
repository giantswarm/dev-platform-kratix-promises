#!/bin/bash

RED='\033[0;31m'
GREEN='\033[0;32m'
BOLD='\033[1m'
NC='\033[0m' # No Color

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/../build-config.sh"
source "$my_dir/../bash_libs/dir-tools.sh"

function test_image() {
  if [[ $# != 1 ]]; then
    echo "Usage: $0 [dir]"
    exit 1
  fi
  dir="$1"
  if [[ ! -f "$dir/Dockerfile" ]]; then
    return
  fi

  echo "* Testing pipeline in directory $dir"
  cd "$dir" || exit 2

  image_name="$images_registry/$images_repo/$(basename "$dir")"
  for c in tests/*; do
    echo -e "** ${BOLD}Test case $c${NC}"
    echo "*** Running test..."
    rm -f "$c/actual/output/*"
    rm -f "$c/actual/metadata/*"
    docker run -it --rm \
      -u "$(id -u):$(id -g)" \
      -v "${PWD}/$c/input:/kratix/input" \
      -v "${PWD}/$c/actual/output/:/kratix/output" \
      -v "${PWD}/$c/actual/metadata:/kratix/metadata" \
      -e TEST_RUN=true \
      "$image_name"
    exit_status=$?
    expected_exit_status=0
    if [[ -f "$c/expected/exitcode" ]]; then
      expected_exit_status=$(cat "$c/expected/exitcode")
    fi
    echo "*** Asserting test's exit code..."
    if [[ $exit_status != "$expected_exit_status" ]]; then
      echo -e "$RED*** Test $c failed with exit code $exit_status when $expected_exit_status was expected.$NC"
      exit 2
    fi

    echo "*** Asserting test's 'output' directory..."
    if ! diff -x '.*' -r "$c/expected/output/" "$c/actual/output/"; then
      echo -e "$RED** Test failed. Test output left for debugging.$NC"
      exit 2
    fi
    if [[ -d "$c/expected/metadata" ]]; then
      echo "*** Asserting test's 'metadata' directory..."
      if ! diff -x '.*' -r "$c/expected/metadata/" "$c/actual/metadata/"; then
        echo -e "$RED** Test failed. Test metadata left for debugging.$NC"
        exit 2
      fi
    fi

    echo -e "$GREEN** Test passed, expected output matches actual output$NC"
  done
  echo -e "$GREEN* All tests passed for $dir$NC"
  echo ""
  cd ../..
}

./build-all.sh
./validate-all.sh
for_dirs "containers" test_image
