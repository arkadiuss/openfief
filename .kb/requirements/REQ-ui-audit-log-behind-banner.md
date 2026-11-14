---
title: Settings' Audit log page must show a banner with the words of the audit-log-behind attention item while that item is open, and no banner otherwise.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:12.688Z'
semantic_text: Settings' Audit log page must show a banner with the words of the audit-log-behind attention item while that item is open, and no banner otherwise.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 26e2887fc1ee6e89a1acdf97ef8423505bebeadda646602eec5a832c5fa03fcf
semantic_inventory:
  - claim_key: CLAIM-178C3FC0060E7B67
    claim_text: Settings' Audit log page must show a banner with the words of the audit-log-behind attention item while that item is open, and no banner otherwise
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 146
logic_claims:
  - CLAIM-178C3FC0060E7B67
id: REQ-ui-audit-log-behind-banner
type: req
---
Settings' Audit log page must show a banner with the words of the audit-log-behind attention item while that item is open, and no banner otherwise.

## Context

The banner is the same state as the Inbox's item, read through the attention items of the read API, so a core that can't list them shows the page without it. It follows the design system's alert component.

## Source

> AUDIT-LOG-BEHIND ATTENTION ITEM: after an audit-entry job has failed 5 attempts (keep River retrying indefinitely; never drop), raise attention item kind audit-log-behind in the data coming in group next to failing syncs (D-196). Cleared when the backlog is empty. Same state as a banner on Settings, Audit log.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18