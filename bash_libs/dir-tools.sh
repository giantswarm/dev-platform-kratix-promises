#!/usr/bin/env bash

function for_dirs() {
  if [[ $# != 2 ]]; then
    echo "Usage: $0 [dir] [function]"
    exit 1
  fi

  for dir in "$1"/*; do
    d=$(basename "$dir")
    if [[ ! -d "$dir" || "$d" == _* ]]; then
      continue
    fi
    (
      "$2" "$dir"
    )
  done
}

function run_script_in_dir() {
  if [[ $# != 2 ]]; then
    echo "Usage: $0 [dir]"
    exit 1
  fi
  if [[ ! -x "$1/$script.sh" ]]; then
    return
  fi
  (
    echo "$script in $1"
    cd "$1" || exit
    ./"$script".sh
  )
}

function run_script_for_dirs() {
  if [[ $# != 2 ]]; then
    echo "Usage: $0 [dir] [script name]"
    exit 1
  fi

  script="$2"
  for_dirs "$1" run_script_in_dir
}
