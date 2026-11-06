---
title: The architecture tests must fail when an import breaks the context map or the layering of domain, app and adapter packages, when a plugin or the SDK imports internal/ or cmd/, or when a package is named util, common, helpers, models or types.
status: open
text_ref: docs/dev/code-organization.md#architecture-tests
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-architecture-tests
  - lane:strict
semantic_text: The architecture tests must fail when an import breaks the context map or the layering of domain, app and adapter packages, when a plugin or the SDK imports internal/ or cmd/, or when a package is named util, common, helpers, models or types.
logic_claims:
  - CLAIM-C05DCE2BC7B8DF43
semantic_clauses:
  - The architecture tests must fail when an import breaks the context map or the layering of domain, app and adapter packages, when a plugin or the SDK imports internal/ or cmd/, or when a package is named util, common, helpers, models or types.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 142fa54344bf3add84910ecc5a391bbaf40211cb07bb541acec69e92d56beaac
semantic_inventory:
  - claim_key: CLAIM-C05DCE2BC7B8DF43
    claim_text: The architecture tests must fail when an import breaks the context map or the layering of domain, app and adapter packages, when a plugin or the SDK imports internal/ or cmd/, or when a package is named util, common, helpers, models or types
    role: normative
    status: modeled
    span:
      start: 0
      end: 241
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:46.529Z'
id: REQ-tooling-archtest-imports
type: req
---
The architecture tests must fail when an import breaks the context map or the layering of domain, app and adapter packages, when a plugin or the SDK imports internal/ or cmd/, or when a package is named util, common, helpers, models or types.

## Context

code-organization.md states the context map, the layering, the plugin rule and the banned package names, and says an architecture test enforces them; depguard matches import prefixes only, so it cannot see which context or layer a package is in. The architecture tests in tools/archtest run with go test in task check and in the pre-commit hook. Implemented by Claude Code for the architecture-tests task the orchestrating agent set after the alpha.1 tag; the author has not reviewed this wording yet. 

## Source

> No context imports another that the map doesn't name, and kernel imports no context.

docs/dev/code-organization.md#architecture-tests
