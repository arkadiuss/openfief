---
title: The release bundle must run the core with --authz=spicedb, so SpiceDB answers authorization on every install, and the bundle script must refuse to render an install whose core runs no --authz or another one.
status: open
firmness: default
required_evidence: unverified
semantic_text: The release bundle must run the core with --authz=spicedb, so SpiceDB answers authorization on every install, and the bundle script must refuse to render an install whose core runs no --authz or another one.
rationale: The author decided that SpiceDB becomes the default authorizer before the 2026.11.0-alpha.2 cut and that the release bundle runs it without option (relayed by the design session).
origin:
  kind: agent
  ref: 'the design session''s task for the SpiceDB default, relaying the author''s decision on D-164, 2026-11-22: Release bundle. The release bundle runs --authz=spicedb, and this is not optional.'
  recorded_at: '2026-11-22T08:41:10.456Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 35089865f4077047613dd6acfe5a792f5e04b4c893e671fd41637d53789c866e
semantic_inventory:
  - claim_key: CLAIM-CC54827DF8BF72DE
    claim_text: The release bundle must run the core with --authz=spicedb, so SpiceDB answers authorization on every install
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 108
    subjects:
      - tooling.release_bundle_authz
  - claim_key: CLAIM-AF1A012055D42CD3
    claim_text: the bundle script must refuse to render an install whose core runs no --authz or another one
    role: normative
    status: ontology_gap
    span:
      start: 114
      end: 206
    subjects:
      - tooling.release_bundle_authz
logic_claims:
  - CLAIM-CC54827DF8BF72DE
  - CLAIM-AF1A012055D42CD3
id: REQ-tooling-release-bundle-authz-spicedb
type: req
---
The release bundle must run the core with --authz=spicedb, so SpiceDB answers authorization on every install, and the bundle script must refuse to render an install whose core runs no --authz or another one.

## Context

The design session relayed the author's decision on 2026-11-22 that SpiceDB becomes the default authorizer before the 2026.11.0-alpha.2 cut (D-164; D-163 and D-164 stay as written), and that in the release bundle it is not optional. deploy/core/deployment.yaml spells out --authz=spicedb, the binary's default too, and scripts/release-bundle.sh's check_authz refuses an install whose core container runs --authz=policy, any other --authz or none. An owner may still opt out on their own install with --authz=policy (docs/guides/install.md, Run without SpiceDB); a release never ships it. Reason not stated beyond the decision itself.

## Source

> Release bundle. The release bundle runs --authz=spicedb, and this is not optional.

Source: the design session's task for the SpiceDB default, relaying the author's decision on D-164, 2026-11-22.
