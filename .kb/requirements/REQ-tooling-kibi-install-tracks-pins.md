---
title: task kibi:install must reinstall the pinned kibi packages whenever an installed kibi version or an applied patch differs from package-lock.json and patches/, and the post-checkout and post-merge hooks must run it when package.json, package-lock.json or patches/ changed.
status: open
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - provenance:mcp-kibi-model-requirement
  - lane:strict
text_ref: docs/dev/pipeline.md#git-hooks
origin:
  kind: agent
  recorded_at: '2026-11-13T21:41:59.613Z'
semantic_text: task kibi:install must reinstall the pinned kibi packages whenever an installed kibi version or an applied patch differs from package-lock.json and patches/, and the post-checkout and post-merge hooks must run it when package.json, package-lock.json or patches/ changed.
logic_claims:
  - CLAIM-9188A450E3D228CA
semantic_clauses:
  - task kibi:install must reinstall the pinned kibi packages whenever an installed kibi version or an applied patch differs from package-lock.json and patches/, and the post-checkout and post-merge hooks must run it when package.json, package-lock.json or patches/ changed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 10980b74258eb47c2599b7deb68c393095577fc9faf13855e550f1a3880a4091
semantic_inventory:
  - claim_key: CLAIM-9188A450E3D228CA
    claim_text: task kibi:install must reinstall the pinned kibi packages whenever an installed kibi version or an applied patch differs from package-lock.json and patches/, and the post-checkout and post-merge hooks must run it when package.json, package-lock.json or patches/ changed
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
id: REQ-tooling-kibi-install-tracks-pins
type: req
---
task kibi:install must reinstall the pinned kibi packages whenever an installed kibi version or an applied patch differs from package-lock.json and patches/, and the post-checkout and post-merge hooks must run it when package.json, package-lock.json or patches/ changed.

## Context

After a merge moved kibi-cli and kibi-core to new pins and changed the patches, task kibi:install said it was up to date, because its sources and generates checksum did not notice; node_modules kept kibi 2.36, and the docs coverage pages took over two minutes and failed the hook until task --force kibi:install. The task now compares the installed kibi versions with package-lock.json and the patches with a stamp the last install wrote (scripts/kibi-install-status.sh), and lefthook runs it after a branch checkout or a merge that changed the pins or patches. Requested by the coordinating agent after another session reported the failure; modeled by Claude Code, the author has not reviewed the wording yet.

## Source

> kibi:install needs a status that fails when any installed kibi package's version differs from the pinned version or the patches are not applied; add post-merge and post-checkout lefthook commands that run task kibi:install only when package.json, package-lock.json or patches/ changed.

docs/dev/pipeline.md#git-hooks
