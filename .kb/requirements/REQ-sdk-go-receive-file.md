---
title: The Go SDK's ReceiveFile must take exactly one header as the first message of an ImportFile stream, then the file's chunks in order, and must refuse a stream without a header, with a second header, or with more bytes than the connector's limit.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: The Go SDK's ReceiveFile must take exactly one header as the first message of an ImportFile stream, then the file's chunks in order, and must refuse a stream without a header, with a second header, or with more bytes than the connector's limit.
logic_claims:
  - CLAIM-A8E67C08803D1EF7
semantic_clauses:
  - The Go SDK's ReceiveFile must take exactly one header as the first message of an ImportFile stream, then the file's chunks in order, and must refuse a stream without a header, with a second header, or with more bytes than the connector's limit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f7d89059e9fbb9a9bfaa535e698a2e51180db6378d106d8eb727192515f50847
semantic_inventory:
  - claim_key: CLAIM-A8E67C08803D1EF7
    claim_text: The Go SDK's ReceiveFile must take exactly one header as the first message of an ImportFile stream, then the file's chunks in order, and must refuse a stream without a header, with a second header, or with more bytes than the connector's limit
    role: normative
    status: modeled
    span:
      start: 0
      end: 243
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:47.855Z'
id: REQ-sdk-go-receive-file
type: req
---
The Go SDK's ReceiveFile must take exactly one header as the first message of an ImportFile stream, then the file's chunks in order, and must refuse a stream without a header, with a second header, or with more bytes than the connector's limit.

## Context

D-49 routes file uploads through the core to the right connector, so a connector receives files as a stream; the limit bounds what a malicious or broken statement can make it hold in memory (threat: malicious statement files, DESIGN.md 3.5). Refines REQ-ingest-core-calls-connectors. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md D-49
