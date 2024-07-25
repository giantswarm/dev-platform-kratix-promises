#!/bin/bash -e

my_dir=$(dirname -- "$(readlink -f -- "$0")")
source "$my_dir/bash_libs/dir-tools.sh"

run_script_for_dirs "$my_dir" build-all
