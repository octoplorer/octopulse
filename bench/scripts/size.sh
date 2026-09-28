#!/usr/bin/env bash
# Report uncompressed size and content hash of each probe binary so a run can be
# compared with another. Avoids stat (its flags differ between BSD and GNU
# userlands) and accepts either sha256sum (coreutils) or shasum (perl, the macOS
# default). Full hashes go to dist/SHA256SUMS for later comparison.
set -euo pipefail
cd "$(dirname "$0")/.."

cd dist
: > SHA256SUMS
for f in base-linux-arm64 combo-linux-arm64 base-linux-amd64 combo-linux-amd64; do
  [ -f "$f" ] || { echo "missing $f (run build.sh first)" >&2; exit 1; }
  bytes=$(wc -c < "$f" | tr -d ' ')
  if command -v sha256sum >/dev/null 2>&1; then
    hash=$(sha256sum "$f" | cut -d' ' -f1)
  else
    hash=$(shasum -a 256 "$f" | cut -d' ' -f1)
  fi
  printf '%-22s %10s bytes  %s\n' "$f" "$bytes" "${hash:0:8}"
  echo "$hash  $f" >> SHA256SUMS
done
echo "full hashes written to dist/SHA256SUMS"
