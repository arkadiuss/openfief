---
title: The core, cmd/core and the packages below it, tests included, never imports internal/views/ui, so the core links none of the page's templates, assets or design system; what the core and the page share, the Ask box's answerers, is internal/views/app/answer.
status: open
firmness: default
origin:
  kind: agent
  ref: Design session decision on the MCP toolbox seam, 2026-11-18
  recorded_at: '2026-11-19T11:20:13.086Z'
id: REQ-tooling-archtest-core-no-ui
type: req
semantic_text: The core, cmd/core and the packages below it, tests included, never imports internal/views/ui, so the core links none of the page's templates, assets or design system; what the core and the page share, the Ask box's answerers, is internal/views/app/answer.
semantic_source_field: semantic_text
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_hash: bb426b1b804384a7d781170d3cf73ae9de52a751941e28ea700fcae4be0f973e
semantic_inventory:
  - claim_key: CLAIM-17EE6E5C79817CE8
    claim_text: The core, cmd/core and the packages below it, tests included, never imports internal/views/ui, so the core links none of the page's templates, assets or design system
    role: descriptive
    status: missing
    span:
      start: 0
      end: 166
  - claim_key: CLAIM-631FD0508E49B9CF
    claim_text: what the core and the page share, the Ask box's answerers, is internal/views/app/answer
    role: descriptive
    status: missing
    span:
      start: 168
      end: 255
logic_claims:
  - CLAIM-17EE6E5C79817CE8
  - CLAIM-631FD0508E49B9CF
---
The core, cmd/core and the packages below it, tests included, never imports internal/views/ui, so the core links none of the page's templates, assets or design system; what the core and the page share, the Ask box's answerers, is internal/views/app/answer.

## Context

The core imported the UI package only to build the MCP server toolbox from the Ask box tools, and its image had to carry the design system for it. The design session decided the answerers move to a package both import and that an architecture test keeps the core from importing the UI again.

## Source

> Design session, 2026-11-18: move the Ask box six read-only answerers and the Toolbox into a neutral package, then add the architecture rule that cmd/core must not import internal/views/ui.
