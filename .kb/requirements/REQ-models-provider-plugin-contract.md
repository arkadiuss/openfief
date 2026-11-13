---
title: A model-provider plugin holds its provider's key and may reach only its provider's host; it declares where its models run and the provider's retention terms, lists models, generates with tools and structured output, and reports each call's usage.
status: open
text_ref: design-decisions:D-205#2
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-205#2
  recorded_at: '2026-11-13T11:28:51.094Z'
semantic_text: A model-provider plugin holds its provider's key and may reach only its provider's host; it declares where its models run and the provider's retention terms, lists models, generates with tools and structured output, and reports each call's usage.
semantic_source_field: semantic_text
semantic_source_hash: 057eb1f55893d2c7685a96b1eb60419a845a79924eeb4f0bc7f5477c95a63ccf
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-FE952A483221A739
    claim_text: A model-provider plugin holds its provider's key and may reach only its provider's host
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 87
  - claim_key: CLAIM-4930667E5440D10F
    claim_text: it declares where its models run and the provider's retention terms, lists models, generates with tools and structured output, and reports each call's usage
    role: descriptive
    status: missing
    span:
      start: 89
      end: 245
logic_claims:
  - CLAIM-FE952A483221A739
  - CLAIM-4930667E5440D10F
semantic_clauses:
  - A model-provider plugin holds its provider's key and may reach only its provider's host
  - it declares where its models run and the provider's retention terms, lists models, generates with tools and structured output, and reports each call's usage
required_evidence: asserted
id: REQ-models-provider-plugin-contract
type: req
---
A model-provider plugin holds its provider's key and may reach only its provider's host; it declares where its models run and the provider's retention terms, lists models, generates with tools and structured output, and reports each call's usage.

## Context

The contract is proto/openfief/models/v1/provider.proto, served by the plugin to the core only; the plugin never sees a role, a data tier or who asked.

## Source

> A plugin holds its provider's key and may reach only its provider's host (D-32); it declares where its models run and the provider's retention terms, lists models, generates with tools and structured output, and reports each call's usage.

Source: OpenFief author decisions (DESIGN.md section 10) - D-205#2