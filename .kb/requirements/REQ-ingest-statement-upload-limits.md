---
title: An uploaded statement must be refused before a connector reads it when the caller is not the owner, no mode is given or the file is empty, and stopped when it passes 64 MiB; the connector gets only the cleaned last element of its name and a sniffed media type.
status: open
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
semantic_text: An uploaded statement must be refused before a connector reads it when the caller is not the owner, no mode is given or the file is empty, and stopped when it passes 64 MiB; the connector gets only the cleaned last element of its name and a sniffed media type.
logic_claims:
  - CLAIM-C2906C4F54F95FE2
semantic_clauses:
  - An uploaded statement must be refused before a connector reads it when the caller is not the owner, no mode is given or the file is empty, and stopped when it passes 64 MiB; the connector gets only the cleaned last element of its name and a sniffed media type.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 854ad344e7a1def566b18ac9521e7474cc064e16ec2d85dbeff993ee0228260b
semantic_inventory:
  - claim_key: CLAIM-C2906C4F54F95FE2
    claim_text: An uploaded statement must be refused before a connector reads it when the caller is not the owner, no mode is given or the file is empty, and stopped when it passes 64 MiB; the connector gets only the cleaned last element of its name and a sniffed media type
    role: normative
    status: modeled
    span:
      start: 0
      end: 259
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:27.828Z'
id: REQ-ingest-statement-upload-limits
type: req
---
An uploaded statement must be refused before a connector reads it when the caller is not the owner, no mode is given or the file is empty, and stopped when it passes 64 MiB; the connector gets only the cleaned last element of its name and a sniffed media type.

## Context

DESIGN 3.7 lists malicious statement files as a threat. 64 MiB is the IBKR Flex parser's own cap, the largest any connector takes today; the browser's file name and media type are hints, never trusted. The limit and the sniffing are this task's choices. Modeled by Claude Code for the statement upload task; the author has not reviewed it yet.

## Source

> **[D-49]** The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches. Views are clients of the read API. File uploads and manual bond-lot entries in the built-in UI go to the core, which hands them to the right connector.

docs/DESIGN.md#D-49
