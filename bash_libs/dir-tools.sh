function run_script_for_dirs() {
	if [[ $# != 2 ]]; then
		echo "Usage: $0 [dir] [script name]"
		exit 1
	fi

	for d in "$1"/*; do
		if [[ ! -d "$d" || ! -x "$d/$2.sh" || "$d" == _* ]]; then
			continue
		fi
		(
			echo "$2 all in $d"
			cd "$d" || exit
			./"$2".sh
		)
	done
}
