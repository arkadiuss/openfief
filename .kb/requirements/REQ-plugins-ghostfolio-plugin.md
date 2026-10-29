---
title: The Ghostfolio plugin's manifest must ask only for the view capability and the ghostfolio-credentials Secret holding the Ghostfolio security token; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach Ghostfolio only through the egress proxy and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
status: open
text_ref: docs/DESIGN.md#D-3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-3
  - lane:strict
semantic_text: The Ghostfolio plugin's manifest must ask only for the view capability and the ghostfolio-credentials Secret holding the Ghostfolio security token; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach Ghostfolio only through the egress proxy and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
logic_claims:
  - CLAIM-A4F7C0D5B8862902
semantic_clauses:
  - The Ghostfolio plugin's manifest must ask only for the view capability and the ghostfolio-credentials Secret holding the Ghostfolio security token; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach Ghostfolio only through the egress proxy and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bae0134dca45629a1b8155961fa1a09ffb4bef5d34f3b456dc9dc46ad8b0b830
semantic_inventory:
  - claim_key: CLAIM-A4F7C0D5B8862902
    claim_text: The Ghostfolio plugin's manifest must ask only for the view capability and the ghostfolio-credentials Secret holding the Ghostfolio security token; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach Ghostfolio only through the egress proxy and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts
    role: normative
    status: modeled
    span:
      start: 0
      end: 541
origin:
  kind: agent
  recorded_at: '2026-10-29T01:37:01.053Z'
id: REQ-plugins-ghostfolio-plugin
type: req
---
The Ghostfolio plugin's manifest must ask only for the view capability and the ghostfolio-credentials Secret holding the Ghostfolio security token; the plugin must refuse a config whose url is not an https origin, whose interval is under five minutes or that lists an invalid or repeated ISIN; must read the read API with its own projected token, reach Ghostfolio only through the egress proxy and read the security token from the Secret on every login; and must sync at start and then every interval, an hour by default, logging only counts.

## Context

D-3 ships Ghostfolio auto-push in M0 and D-4 makes it a reconcile loop, so the plugin runs the loop on its own timer: views are clients of the read API (D-49), and the core schedules only connectors. The manifest is the signed image label (D-42), so it asks for no more than the plugin needs; the owner's Ghostfolio host can't be in a label built before the owner chose it, so the manifest names no egress host (an open question for the author). Credentials are a Secret the owner creates (D-29); logs carry no amount (GR-12). Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

## Source

> M0 ships two view plugins: a very simple built-in UI as the default, and Ghostfolio auto-push.

docs/DESIGN.md D-3

> The Ghostfolio plugin runs a reconcile loop. It owns only the Ghostfolio accounts it created, puts OpenFief's transaction ID in each activity's comment, deletes stale activities and imports missing ones. Manual edits inside its accounts are reverted on the next sync.

docs/DESIGN.md D-4
