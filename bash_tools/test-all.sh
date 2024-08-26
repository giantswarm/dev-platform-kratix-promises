#!/bin/bash

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/../bash_libs/test-all.sh"

./build-all.sh
./validate-all.sh
test_dir "containers"
