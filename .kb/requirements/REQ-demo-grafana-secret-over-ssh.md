---
title: task demo:grafana-secret creates or replaces the Secret grafana-otlp in the namespace observability, with the keys endpoint, instance-id and token, by reading the three values locally (the token without echo) and sending them over SSH on standard input as an env file into kubectl create secret --from-env-file --dry-run=client | kubectl apply, then restarting the collector, so no value is in an argument, a file, git, CI or a log and none is printed; it refuses without DEMO_SSH=keys, an endpoint that is not an https address on grafana.net, an instance ID that is not a number, a token that does not look like one and a missing value, and running it again rotates the Secret.
status: open
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data (GR-12, D-172, D-203, D-227), 2026-11-19'
  recorded_at: '2026-11-19T16:20:42.639Z'
semantic_text: task demo:grafana-secret creates or replaces the Secret grafana-otlp in the namespace observability, with the keys endpoint, instance-id and token, by reading the three values locally (the token without echo) and sending them over SSH on standard input as an env file into kubectl create secret --from-env-file --dry-run=client | kubectl apply, then restarting the collector, so no value is in an argument, a file, git, CI or a log and none is printed; it refuses without DEMO_SSH=keys, an endpoint that is not an https address on grafana.net, an instance ID that is not a number, a token that does not look like one and a missing value, and running it again rotates the Secret.
logic_claims:
  - CLAIM-CE787EE53987A6B1
  - CLAIM-5EB8DCB3DCBD1A94
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
semantic_source_hash: 8739bc23477ee796d1fc2b7b0c0e999c1c758edbd099e7aa1d2f03e13f88138f
semantic_inventory:
  - claim_key: CLAIM-CE787EE53987A6B1
    claim_text: task demo:grafana-secret creates or replaces the Secret grafana-otlp in the namespace observability, with the keys endpoint, instance-id and token, by reading the three values locally (the token without echo) and sending them over SSH on standard input as an env file into kubectl create secret --from-env-file --dry-run=client | kubectl apply, then restarting the collector, so no value is in an argument, a file, git, CI or a log and none is printed
    role: descriptive
    status: missing
    span:
      start: 0
      end: 451
  - claim_key: CLAIM-5EB8DCB3DCBD1A94
    claim_text: it refuses without DEMO_SSH=keys, an endpoint that is not an https address on grafana.net, an instance ID that is not a number, a token that does not look like one and a missing value, and running it again rotates the Secret
    role: descriptive
    status: missing
    span:
      start: 453
      end: 677
id: REQ-demo-grafana-secret-over-ssh
type: req
---
task demo:grafana-secret creates or replaces the Secret grafana-otlp in the namespace observability, with the keys endpoint, instance-id and token, by reading the three values locally (the token without echo) and sending them over SSH on standard input as an env file into kubectl create secret --from-env-file --dry-run=client | kubectl apply, then restarting the collector, so no value is in an argument, a file, git, CI or a log and none is printed; it refuses without DEMO_SSH=keys, an endpoint that is not an https address on grafana.net, an instance ID that is not a number, a token that does not look like one and a missing value, and running it again rotates the Secret.

## Context

The collector's credentials are the author's and must never reach git or CI, as the OpenRouter key and the Turnstile secret key never do. The endpoint is checked because the collector sends there and nowhere else, and the machine's NetworkPolicy can't name a host.

## Source

> A task demo:grafana-secret modelled on demo:model-secret: it reads the values locally without echoing them and creates the Secret over SSH, never in git or CI (the design session's demo operations brief, 2026-11-19).