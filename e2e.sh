#!/usr/bin/env bash
# End-to-end check: build the container image from the shared Dockerfile and
# feed every testdata/input fixture through it as a Markdown fence. The
# container's output is diffed against testdata/golden. No Go test process runs
# here; the container entry point is the only renderer.
#
# Usage:
#   ./e2e.sh                 # build the image and verify every fixture
#   UPDATE_GOLDEN=1 ./e2e.sh # rewrite golden fixtures from the container output
set -euo pipefail

ROOT=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
IMAGE=${IMAGE:-goldmark-diagrams:test}
BUILD_TIMEOUT=${BUILD_TIMEOUT:-600}
RUN_TIMEOUT=${RUN_TIMEOUT:-120}

if ! docker info >/dev/null 2>&1; then
	echo "e2e: docker is not available" >&2
	exit 1
fi

timeout "$BUILD_TIMEOUT" "$ROOT/build.sh"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

fail=0
count=0
for input in "$ROOT"/testdata/input/*; do
	name=$(basename "$input")
	name=${name%.*}
	lang=${input##*.}
	golden="$ROOT/testdata/golden/$name.html"
	source=$(cat "$input")

	printf '```%s\n%s\n```\n' "$lang" "$source" > "$work/input.md"
	if ! timeout "$RUN_TIMEOUT" docker run --rm -i "$IMAGE" < "$work/input.md" > "$work/output.html"; then
		echo "e2e: $name: container failed" >&2
		fail=1
		continue
	fi
	if [ -n "${UPDATE_GOLDEN:-}" ]; then
		cp "$work/output.html" "$golden"
		count=$((count + 1))
		continue
	fi
	if ! diff -u "$golden" "$work/output.html" > "$work/diff"; then
		echo "e2e: $name: output differs from $golden" >&2
		head -40 "$work/diff" >&2
		fail=1
		continue
	fi
	count=$((count + 1))
done

if [ "$fail" -ne 0 ]; then
	exit 1
fi
if [ -n "${UPDATE_GOLDEN:-}" ]; then
	echo "e2e: rewrote $count golden fixtures"
else
	echo "e2e: $count fixtures match"
fi