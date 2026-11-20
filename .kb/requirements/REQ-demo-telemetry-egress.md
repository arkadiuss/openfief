---
title: The demo's telemetry collector sends only to the one exporter otlphttp/grafana, whose endpoint, instance ID and token come from the Secret grafana-otlp and are in no file in git, its NetworkPolicy admits only the core and the operator to its OTLP port 4318 and lets it out only to DNS and to HTTPS on the public internet with the private, shared, link-local and loopback ranges left out, it mounts the node's pod logs read-only, and without the Secret its pod waits without running; scripts/demo/validate.sh checks the built layer.
status: open
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data (GR-12, D-172, D-203, D-227), 2026-11-19'
  recorded_at: '2026-11-19T16:20:42.098Z'
semantic_text: The demo's telemetry collector sends only to the one exporter otlphttp/grafana, whose endpoint, instance ID and token come from the Secret grafana-otlp and are in no file in git, its NetworkPolicy admits only the core and the operator to its OTLP port 4318 and lets it out only to DNS and to HTTPS on the public internet with the private, shared, link-local and loopback ranges left out, it mounts the node's pod logs read-only, and without the Secret its pod waits without running; scripts/demo/validate.sh checks the built layer.
logic_claims:
  - CLAIM-E4087BAE2EECCA31
  - CLAIM-C57439E0BC5B4C3A
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
semantic_source_hash: c5bbb150d41d710083ecc6afed739ce6ad2da257bf436029d9e88e4aa8bb77de
semantic_inventory:
  - claim_key: CLAIM-E4087BAE2EECCA31
    claim_text: The demo's telemetry collector sends only to the one exporter otlphttp/grafana, whose endpoint, instance ID and token come from the Secret grafana-otlp and are in no file in git, its NetworkPolicy admits only the core and the operator to its OTLP port 4318 and lets it out only to DNS and to HTTPS on the public internet with the private, shared, link-local and loopback ranges left out, it mounts the node's pod logs read-only, and without the Secret its pod waits without running
    role: descriptive
    status: missing
    span:
      start: 0
      end: 481
  - claim_key: CLAIM-C57439E0BC5B4C3A
    claim_text: scripts/demo/validate.sh checks the built layer
    role: descriptive
    status: missing
    span:
      start: 483
      end: 530
id: REQ-demo-telemetry-egress
type: req
---
The demo's telemetry collector sends only to the one exporter otlphttp/grafana, whose endpoint, instance ID and token come from the Secret grafana-otlp and are in no file in git, its NetworkPolicy admits only the core and the operator to its OTLP port 4318 and lets it out only to DNS and to HTTPS on the public internet with the private, shared, link-local and loopback ranges left out, it mounts the node's pod logs read-only, and without the Secret its pod waits without running; scripts/demo/validate.sh checks the built layer.

## Context

A NetworkPolicy can't select by hostname, so what keeps the collector to Grafana is its one exporter, whose address is the Secret's, the NetworkPolicy's narrowing to HTTPS and the private ranges, and the check that nothing else is configured. The pod is not a plugin, so it does not use the egress proxy. A missing Secret keeps the pod waiting in CreateContainerConfigError: it runs nothing and logs nothing, instead of crash-looping.

## Source

> The collector's own NetworkPolicy allows egress only to DNS and the Grafana OTLP host, port 443; ingress only from the core; if the Secret is absent the collector must not crash-loop noisily (the design session's demo operations brief, 2026-11-19).