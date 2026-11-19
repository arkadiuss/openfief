#!/usr/bin/env bash
# Creates or replaces one of the demo's Secrets out of band (docs/dev/demo.md#the-demo-model). Called through task:
#   task demo:model-secret ENV=demo        the OpenRouter key: Secret openrouter--api-key, file api-key
#   task demo:turnstile-secret ENV=demo    the Turnstile secret key: Secret ui--turnstile-secret, file secret-key
#   task demo:grafana-secret ENV=demo      the Grafana Cloud OTLP endpoint, instance ID and token: Secret grafana-otlp in the
#                                          namespace observability, keys endpoint, instance-id and token, which the telemetry
#                                          layer's collector reads (docs/dev/demo.md#telemetry)
# The value is read locally, with `read -rs` on a terminal (or the first line of a pipe, for a script), and goes over
# SSH on the command's standard input into
#   k3s kubectl -n openfief create secret generic <name> --from-file=<file>=/dev/stdin --dry-run=client -o yaml |
#     k3s kubectl apply -f -
# (for grafana, three key=value lines into --from-env-file=/dev/stdin, in namespace observability)
# so it is in no argument, no environment of the machine, no file, no cloud-init, no git and no log, and nothing prints
# it. Running it again rotates the value (create or replace). It needs SSH to the machine: DEMO_SSH=keys, with
# DEMO_SSH_CIDRS holding your address when the machine was created or last brought up, and the firewall open to it
# (docs/dev/demo.md#ssh). The Hetzner token (HCLOUD_TOKEN, or HETZNER_API_TOKEN) only finds the machine's address.
set -euo pipefail
shopt -s inherit_errexit

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# demo.sh runs only when executed, so sourcing it gives the pure helpers (demo_check_env, demo_fail).
# shellcheck source=scripts/demo/demo.sh
source "$here/demo.sh"

# demo_secret_spec prints "<variable> <secret name> <file key> <what>" for the Secret a command creates, or fails.
demo_secret_spec() {
  case "$1" in
    model) echo "OPENROUTER_API_KEY openrouter--api-key api-key the-OpenRouter-key" ;;
    turnstile) echo "TURNSTILE_SECRET_KEY ui--turnstile-secret secret-key the-Turnstile-secret-key" ;;
    *)
      demo_fail "usage: ENV=<name> DEMO_SSH=keys secret.sh model|turnstile|grafana"
      return 1
      ;;
  esac
}

# demo_secret_read prints the value of the variable named by the first argument: its value in the environment, else
# read without echo from the terminal (or a pipe), else it fails; a third argument, visible, reads it with echo, for a
# value that is no secret, such as an address. The value goes to stdout only, for the caller to hold in a variable,
# never in an argument.
demo_secret_read() {
  local name="$1" what="$2" mode="${3:-}" value=''
  if [ -n "${!name:-}" ]; then
    value="${!name}"
  elif [ -t 0 ] && [ "$mode" = visible ]; then
    read -r -p "$what: " value >&2 || true
  elif [ -t 0 ]; then
    read -rs -p "$what (not echoed): " value >&2 || true
    echo >&2
  else
    IFS= read -r value || true
  fi
  if [ -z "$value" ]; then
    demo_fail "no value for $what: set $name, or type it when asked"
    return 1
  fi
  if [[ "$value" == *[[:space:]]* ]]; then
    demo_fail "the value has white space in it, which a key or secret never has: not sent"
    return 1
  fi
  printf '%s' "$value"
}

# demo_secret_host prints the machine's IPv4 address for the environment, from Hetzner.
demo_secret_host() {
  if [ -z "${HCLOUD_TOKEN:-}" ] && [ -n "${HETZNER_API_TOKEN:-}" ]; then
    HCLOUD_TOKEN="$HETZNER_API_TOKEN"
  fi
  if [ -z "${HCLOUD_TOKEN:-}" ]; then
    demo_fail "HCLOUD_TOKEN is not set (nor HETZNER_API_TOKEN): it finds the machine's address; read -rs HCLOUD_TOKEN && export HCLOUD_TOKEN"
    return 1
  fi
  export HCLOUD_TOKEN
  # A missing config file: hcloud uses HCLOUD_TOKEN, never a stored context of another project.
  export HCLOUD_CONFIG="${work:?}/no-cli.toml"
  unset HCLOUD_CONTEXT
  local ip
  ip="$(hcloud server ip "openfief-$ENV")"
  if [[ ! "$ip" =~ ^[0-9]{1,3}(\.[0-9]{1,3}){3}$ ]]; then
    demo_fail "no IPv4 address for server openfief-$ENV"
    return 1
  fi
  printf '%s' "$ip"
}

# demo_secret_connect checks the environment and that SSH is on, and prints the machine's IPv4 address.
demo_secret_connect() {
  local tool
  demo_check_env "${ENV:-}"
  if [ "${DEMO_SSH:-}" != keys ]; then
    demo_fail "this needs SSH to the machine: run it with DEMO_SSH=keys, after task demo:up ENV=$ENV DEMO_SSH_CIDRS=<your address>/32
opened the firewall to you. When it is done, task demo:up ENV=$ENV DEMO_SSH_CIDRS= closes SSH again"
    return 1
  fi
  for tool in hcloud ssh; do
    command -v "$tool" >/dev/null 2>&1 || demo_fail "$tool not found" || return 1
  done
  demo_secret_host
}

# demo_grafana_values reads the three values of the Secret grafana-otlp and prints them as the key=value lines of an env
# file, endpoint first, or fails without printing a value: the endpoint must be an https address on grafana.net (the
# collector sends there and nowhere else, and the machine's NetworkPolicy can't say so), the instance ID a number, and the
# token one token.
demo_grafana_values() {
  local endpoint instance token
  endpoint="$(demo_secret_read GRAFANA_OTLP_ENDPOINT "the Grafana OTLP endpoint (https://otlp-gateway-<zone>.grafana.net/otlp)" visible)"
  instance="$(demo_secret_read GRAFANA_INSTANCE_ID "the Grafana instance ID (a number)" visible)"
  token="$(demo_secret_read GRAFANA_OTLP_TOKEN "the Grafana OTLP token")"
  endpoint="${endpoint%/}"
  if [[ ! "$endpoint" =~ ^https://[a-z0-9-]+([.][a-z0-9-]+)*[.]grafana[.]net(/[A-Za-z0-9._/-]*)?$ ]]; then
    demo_fail "GRAFANA_OTLP_ENDPOINT must be an https address on grafana.net, such as https://otlp-gateway-prod-eu-west-0.grafana.net/otlp (the stack's page in Grafana Cloud, under OpenTelemetry, shows it): not sent"
    return 1
  fi
  if [[ ! "$instance" =~ ^[0-9]{1,20}$ ]]; then
    demo_fail "GRAFANA_INSTANCE_ID must be the stack's number: not sent"
    return 1
  fi
  if [[ ! "$token" =~ ^[A-Za-z0-9_.:+/=-]{8,}$ ]]; then
    demo_fail "GRAFANA_OTLP_TOKEN does not look like a token (letters, digits and _ . : + / = - only, at least 8): not sent"
    return 1
  fi
  printf 'endpoint=%s\ninstance-id=%s\ntoken=%s\n' "$endpoint" "$instance" "$token"
}

# demo_grafana_main creates or replaces the Secret grafana-otlp in namespace observability from the three values, on the
# standard input of one ssh, and restarts the collector so it reads them. A restart that finds no collector yet, because
# Flux has not applied the layer, changes nothing: the pod starts by itself once the Secret exists.
demo_grafana_main() {
  local ip values
  ip="$(demo_secret_connect)"
  values="$(demo_grafana_values)"
  printf '%s\n' "$values" | ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new "root@$ip" \
    "k3s kubectl -n observability create secret generic grafana-otlp --from-env-file=/dev/stdin --dry-run=client -o yaml | k3s kubectl apply -f - && { k3s kubectl -n observability rollout restart daemonset/otel-collector || true; }"
  values=''
  echo "demo '$ENV': the Secret grafana-otlp is created or replaced in the namespace observability, and the collector restarts to read it; nothing was printed or written."
}

# demo_secret_main reads the value and creates or replaces the Secret of command (model or turnstile).
demo_secret_main() {
  local command="${1:-}" spec variable name key what ip value
  spec="$(demo_secret_spec "$command")"
  read -r variable name key what <<<"$spec"
  ip="$(demo_secret_connect)"
  value="$(demo_secret_read "$variable" "${what//-/ }")"
  # The remote command names the Secret and its file and holds no value. --dry-run=client | apply creates the Secret
  # or replaces it, so a second run rotates it.
  printf '%s' "$value" | ssh -o BatchMode=yes -o StrictHostKeyChecking=accept-new "root@$ip" \
    "k3s kubectl -n openfief create secret generic $name --from-file=$key=/dev/stdin --dry-run=client -o yaml | k3s kubectl apply -f -"
  value=''
  echo "demo '$ENV': the Secret $name is created or replaced in the namespace openfief; nothing was printed or written. A plugin that reads it on each call (the model's key) uses it at once; one that reads it at start needs its pod restarted."
}

if [ "${BASH_SOURCE[0]}" = "$0" ]; then
  work="$(mktemp -d)"
  trap 'rm -rf "$work"' EXIT
  if [ "${1:-}" = grafana ]; then
    demo_grafana_main
  else
    demo_secret_main "${1:-}"
  fi
fi
