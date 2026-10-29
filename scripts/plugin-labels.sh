#!/usr/bin/env bash
# Keeps each first-party plugin's signed manifest label in step with its manifest (D-42). The source of truth is
# plugins/<name>/manifest.json; plugins/<name>/Dockerfile carries it, compacted, as the image config label
#   LABEL io.openfief.plugin.manifest='<json>'
# and a Plugin for that image under deploy/demo/plugins must declare the same manifest in spec.manifest, since the
# operator refuses a Plugin whose manifest differs from its image's label.
# `write` rewrites the Dockerfile labels, adding one at the end of a Dockerfile that has none; `check` fails,
# naming each file, when a label or a demo Plugin differs.
# Usage: plugin-labels.sh check|write [repository root]   (gojq on PATH)
set -euo pipefail

label_key='io.openfief.plugin.manifest'

# manifest_label prints a manifest file as the compact JSON its label holds.
manifest_label() {
  gojq -c . "$1"
}

# dockerfile_label prints the manifest label's value in a Dockerfile, without its quotes; nothing when it has none.
dockerfile_label() {
  sed -n "s/^LABEL $label_key='\(.*\)'\$/\1/p" "$1"
}

# write_label replaces the manifest label line of a Dockerfile with value. A Dockerfile without one gets it
# appended, so it lands in the final stage, the image that is signed.
write_label() {
  local dockerfile="$1" value="$2" out
  if [ ! -f "$dockerfile" ]; then
    echo "plugin-labels: $dockerfile is missing" >&2
    return 1
  fi
  # The value goes through the environment: awk -v would interpret its backslashes.
  out="$(LABEL_VALUE="$value" awk -v key="$label_key" '
    index($0, "LABEL " key "=") == 1 { print "LABEL " key "=\047" ENVIRON["LABEL_VALUE"] "\047"; done = 1; next }
    { print }
    END {
      if (!done) {
        print "# The plugin manifest (D-42), signed with the image: written from manifest.json by `task plugins:labels`."
        print "LABEL " key "=\047" ENVIRON["LABEL_VALUE"] "\047"
      }
    }' "$dockerfile")"
  printf '%s\n' "$out" >"$dockerfile"
}

# demo_manifests prints "<image name> <compact sorted manifest>" for each Plugin under deploy/demo/plugins.
demo_manifests() {
  local root="$1" file
  for file in "$root"/deploy/demo/plugins/*.yaml; do
    [ -f "$file" ] || continue
    gojq -r --yaml-input 'select(.kind == "Plugin") | "\(.spec.image | split("@")[0] | split("/")[-1]) \(.spec.manifest | tojson)"' "$file"
  done
}

plugin_labels() {
  local mode="$1" root="${2:-.}" manifest dir name want got failures=0 image declared demo
  case "$mode" in
    check | write) ;;
    *)
      echo "plugin-labels: usage: plugin-labels.sh check|write [root]" >&2
      return 1
      ;;
  esac
  demo="$(demo_manifests "$root")"
  for manifest in "$root"/plugins/*/manifest.json; do
    [ -f "$manifest" ] || continue
    dir="$(dirname "$manifest")"
    name="$(basename "$dir")"
    want="$(manifest_label "$manifest")"
    if [[ "$want" == *"'"* ]]; then
      echo "plugin-labels: $manifest contains a single quote, which the Dockerfile label can't hold" >&2
      failures=$((failures + 1))
      continue
    fi
    if [ "$mode" = write ]; then
      write_label "$dir/Dockerfile" "$want" || failures=$((failures + 1))
      continue
    fi
    got="$(dockerfile_label "$dir/Dockerfile" 2>/dev/null || true)"
    if [ "$got" != "$want" ]; then
      echo "plugin-labels: $dir/Dockerfile's $label_key label differs from $manifest; run task plugins:labels" >&2
      failures=$((failures + 1))
    fi
    while read -r image declared; do
      [ "$image" = "$name" ] || continue
      # gojq prints object keys sorted, so equal manifests print the same.
      declared="$(gojq -c . <<<"$declared")"
      if [ -z "$declared" ] || [ "$declared" != "$want" ]; then
        echo "plugin-labels: a Plugin of $name under deploy/demo/plugins declares another manifest than $manifest" >&2
        failures=$((failures + 1))
      fi
    done <<<"$demo"
  done
  if [ "$failures" -ne 0 ]; then
    return 1
  fi
  echo "plugin-labels: every plugin label matches its manifest"
}

plugin_labels "$@"
