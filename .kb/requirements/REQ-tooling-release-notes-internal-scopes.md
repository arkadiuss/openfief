---
title: Release notes must list a feature or fix whose Conventional Commit scopes all name OpenFief's own tooling or its public demo under Other changes, keeping a scope that also names the product in its section and a breaking change under Breaking changes.
status: open
rationale: 'The design session''s review of the 2026.11.0-alpha.2 notes, 2026-11-21: features and fixes of kibi, the eval bench, the demo, Grafana and the e2e tests crowded Features and Fixes, though nothing an owner installs changes in them.'
semantic_text: Release notes must list a feature or fix whose Conventional Commit scopes all name OpenFief's own tooling or its public demo under Other changes, keeping a scope that also names the product in its section and a breaking change under Breaking changes.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a525246be6a7ef4e965cadcce8321f8cdb7e77739bb6ef9ea7a92d7746c9e2b1
semantic_inventory:
  - claim_key: CLAIM-ACECAC0000AF0267
    claim_text: Release notes must list a feature or fix whose Conventional Commit scopes all name OpenFief's own tooling or its public demo under Other changes, keeping a scope that also names the product in its section and a breaking change under Breaking changes
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 249
    subjects:
      - tooling.release_notes
logic_claims:
  - CLAIM-ACECAC0000AF0267
origin:
  kind: agent
  recorded_at: '2026-11-21T23:06:56.416Z'
id: REQ-tooling-release-notes-internal-scopes
type: req
pair_reviews:
  - with: REQ-tooling-release-notes
    verdict: no_conflict
    separated_by: 'which section a feat or fix lands in when every scope it names is OpenFief''s own tooling or its public demo (this requirement) against: grouping every other commit by its Conventional Commit type and where a release''s notes start; this requirement narrows Features and Fixes and changes neither rule of the other'
    claims:
      - claim_key: CLAIM-ACECAC0000AF0267
    origin:
      kind: agent
      ref: The orchestrator, from the design session's review of the 2026.11.0-alpha.2 notes, 2026-11-21; both requirements have default firmness
---
Release notes must list a feature or fix whose Conventional Commit scopes all name OpenFief's own tooling or its public demo under Other changes, keeping a scope that also names the product in its section and a breaking change under Breaking changes.

## Context

The design session's review of the 2026.11.0-alpha.2 notes, 2026-11-21: features and fixes of kibi, the eval bench, the demo, Grafana and the e2e tests crowded Features and Fixes, though nothing an owner installs changes in them.
