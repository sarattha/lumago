#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output_dir="${repo_root}/captures"
output_path="${output_dir}/phase00-first-frame"

if command -v renderdoccmd >/dev/null 2>&1; then
  renderdoccmd_bin="$(command -v renderdoccmd)"
elif [[ -x "/Applications/RenderDoc.app/Contents/MacOS/renderdoccmd" ]]; then
  renderdoccmd_bin="/Applications/RenderDoc.app/Contents/MacOS/renderdoccmd"
else
  cat >&2 <<'MSG'
RenderDoc CLI was not found.

Install RenderDoc, then rerun this script. On macOS the expected CLI is either:
  - renderdoccmd on PATH
  - /Applications/RenderDoc.app/Contents/MacOS/renderdoccmd
MSG
  exit 1
fi

mkdir -p "${output_dir}"
cd "${repo_root}"
make shaders
"${renderdoccmd_bin}" capture --wait-for-exit --output "${output_path}" -- make run
