---
title: Settings has an Audit log page that shows the owner every entry, newest first and paged, filterable by kind of actor, action, outcome and range of days, naming who, what and which records by name, and no viewer or demo session sees it.
status: open
origin:
  kind: agent
  ref: 'S7 task brief: the Access page''s history and Settings'' audit log view (D-210, D-211), 2026-11-13'
  recorded_at: '2026-11-14T00:30:00.000Z'
firmness: default
semantic_text: Settings has an Audit log page that shows the owner every entry, newest first and paged, filterable by kind of actor, action, outcome and range of days, naming who, what and which records by name, and no viewer or demo session sees it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 61327876fceeef7908c73274a8e083099073ac8f3090fce28f38fa171287e630
semantic_inventory:
  - claim_key: CLAIM-D75E6C68E2736753
    claim_text: Settings has an Audit log page that shows the owner every entry, newest first and paged, filterable by kind of actor, action, outcome and range of days, naming who, what and which records by name, and no viewer or demo session sees it
    role: descriptive
    status: missing
    span:
      start: 0
      end: 234
logic_claims:
  - CLAIM-D75E6C68E2736753
id: REQ-ui-audit-log-page
type: req
---
Settings has an Audit log page that shows the owner every entry, newest first and paged, filterable by kind of actor, action, outcome and range of days, naming who, what and which records by name, and no viewer or demo session sees it.

## Context

D-211 says the log is shown to the owner in Settings, filterable. The brief for the history slice asked for filters by actor kind, action, outcome and date range, paging, links from the Access history, and the owner-only rule. The demo's session is read-only and has no owner session, so, like a viewer, it has no audit log.

## Source

S7 task brief: Settings -> Audit log page (owner-only), D-211 'filterable'; viewers and demo sessions never see either.