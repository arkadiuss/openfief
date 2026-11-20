---
title: task grafana:apply puts deploy/grafana, the folder openfief-demo with its dashboards and alert rules, each under a fixed uid, into the demo's Grafana Cloud stack by upsert, so a second run changes nothing; it refuses a GRAFANA_URL that is missing or is not an https address on grafana.net, a missing GRAFANA_CLOUD_TOKEN and files that fail task grafana:check, before it calls anything, and it prints only names and HTTP statuses, so the token is in no argument, file or output; task grafana:check validates offline that every file is JSON with a fixed uid, that every query uses only the stack's datasources and the metrics the core exports, and that no file holds a secret.
status: open
firmness: default
semantic_text: task grafana:apply puts deploy/grafana, the folder openfief-demo with its dashboards and alert rules, each under a fixed uid, into the demo's Grafana Cloud stack by upsert, so a second run changes nothing; it refuses a GRAFANA_URL that is missing or is not an https address on grafana.net, a missing GRAFANA_CLOUD_TOKEN and files that fail task grafana:check, before it calls anything, and it prints only names and HTTP statuses, so the token is in no argument, file or output; task grafana:check validates offline that every file is JSON with a fixed uid, that every query uses only the stack's datasources and the metrics the core exports, and that no file holds a secret.
origin:
  kind: agent
  ref: 'Author request: the demo''s Grafana Cloud stack as code, git-tracked JSON plus an idempotent apply script (docs/dev/demo.md#grafana), 2026-11-20'
  recorded_at: '2026-11-20T12:32:30.059Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fad05f23ba78952ae07adb5ff421d7b8cb1f1dcf3d4bd09664af5c8ff32dee6a
semantic_inventory:
  - claim_key: CLAIM-8A4E1070609871FD
    claim_text: task grafana:apply puts deploy/grafana, the folder openfief-demo with its dashboards and alert rules, each under a fixed uid, into the demo's Grafana Cloud stack by upsert, so a second run changes nothing
    role: descriptive
    status: missing
    span:
      start: 0
      end: 204
  - claim_key: CLAIM-97224B8296C021C8
    claim_text: it refuses a GRAFANA_URL that is missing or is not an https address on grafana.net, a missing GRAFANA_CLOUD_TOKEN and files that fail task grafana:check, before it calls anything, and it prints only names and HTTP statuses, so the token is in no argument, file or output
    role: normative
    status: ontology_gap
    span:
      start: 206
      end: 476
  - claim_key: CLAIM-89BCBC95915BFBFE
    claim_text: task grafana:check validates offline that every file is JSON with a fixed uid, that every query uses only the stack's datasources and the metrics the core exports, and that no file holds a secret
    role: descriptive
    status: missing
    span:
      start: 478
      end: 673
logic_claims:
  - CLAIM-8A4E1070609871FD
  - CLAIM-97224B8296C021C8
  - CLAIM-89BCBC95915BFBFE
id: REQ-demo-grafana-as-code
type: req
---
task grafana:apply puts deploy/grafana, the folder openfief-demo with its dashboards and alert rules, each under a fixed uid, into the demo's Grafana Cloud stack by upsert, so a second run changes nothing; it refuses a GRAFANA_URL that is missing or is not an https address on grafana.net, a missing GRAFANA_CLOUD_TOKEN and files that fail task grafana:check, before it calls anything, and it prints only names and HTTP statuses, so the token is in no argument, file or output; task grafana:check validates offline that every file is JSON with a fixed uid, that every query uses only the stack's datasources and the metrics the core exports, and that no file holds a secret.

## Context

The orchestrator reads the demo through Grafana Cloud ([Telemetry](docs/dev/demo.md#telemetry)), and what the stack shows and alerts on should be reviewed, versioned and reproducible like the rest of the demo, not clicked together in the UI. The apply token is an Admin service account's, so it must reach only the author's own stack and never git, CI or a log. The author asked for git-tracked JSON and an idempotent script rather than Terraform.

## Source

> Configure the demo's Grafana Cloud stack as code: git-tracked JSON plus an idempotent apply script; require GRAFANA_URL and GRAFANA_CLOUD_TOKEN, refuse a URL that is not https://*.grafana.net, print only names and HTTP statuses (the author's request, 2026-11-20).
