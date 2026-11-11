#!/usr/bin/env bash
set -euo pipefail

vault_path=${1:?usage: deploy.sh <vault-path> [vault-name], run from the plugin project root}
vault_name=${2:-$(basename "$vault_path")}
id=$(node -p 'require("./manifest.json").id')
dest="$vault_path/.obsidian/plugins/$id"

[ -f main.js ] || { echo "main.js missing: run npm run build first" >&2; exit 1; }
mkdir -p "$dest"
cp main.js manifest.json "$dest/"
[ -f styles.css ] && cp styles.css "$dest/"
echo "copied to $dest"

command -v obsidian >/dev/null || { echo "obsidian CLI not on PATH: reload the plugin by hand"; exit 0; }

failed() { grep -q '^Error' <<<"$1"; }
out=$(obsidian vault="$vault_name" plugin:reload id="$id" 2>&1 || true)
if failed "$out"; then
  out=$(obsidian vault="$vault_name" eval code="(async()=>{await app.plugins.loadManifests();await app.plugins.enablePluginAndSave(\"$id\");return \"enabled\"})()" 2>&1 || true)
fi
echo "$out"
! failed "$out"
