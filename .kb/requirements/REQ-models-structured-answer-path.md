---
title: The core asks the Ask box's structured answer strictly of a model whose entry has structured output and in JSON mode of one that has not, through a response schema in the Ask contract and a JSON-mode fallback in the provider contract, and the model-openai plugin sends json_schema or json_object by the entry's flag, refuses a strict request for an entry without it, and asks OpenRouter to require parameters for either format.
status: open
firmness: default
required_evidence: unverified
semantic_text: The core asks the Ask box's structured answer strictly of a model whose entry has structured output and in JSON mode of one that has not, through a response schema in the Ask contract and a JSON-mode fallback in the provider contract, and the model-openai plugin sends json_schema or json_object by the entry's flag, refuses a strict request for an entry without it, and asks OpenRouter to require parameters for either format.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c6481c1e2e0b7b7f5e50cd73a79f436e5a355467c35358ae428f16a3109b2334
semantic_inventory:
  - claim_key: CLAIM-47A3AF09C4E10106
    claim_text: The core asks the Ask box's structured answer strictly of a model whose entry has structured output and in JSON mode of one that has not, through a response schema in the Ask contract and a JSON-mode fallback in the provider contract, and the model-openai plugin sends json_schema or json_object by the entry's flag, refuses a strict request for an entry without it, and asks OpenRouter to require parameters for either format
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 426
logic_claims:
  - CLAIM-47A3AF09C4E10106
origin:
  kind: agent
  ref: 'AB3 brief from the design session: exact Ask box scorers, a structured final answer, host pinning in the bench, 2026-11-14'
  recorded_at: '2026-11-14T14:53:08.547Z'
tags:
  - askbench
id: REQ-models-structured-answer-path
type: req
---
The core asks the Ask box's structured answer strictly of a model whose entry has structured output and in JSON mode of one that has not, through a response schema in the Ask contract and a JSON-mode fallback in the provider contract, and the model-openai plugin sends json_schema or json_object by the entry's flag, refuses a strict request for an entry without it, and asks OpenRouter to require parameters for either format.

## Context

Some models of the route, such as qwen3.7-flash, have no structured-output flag but take a JSON object; the Ask box has to be able to use them. The design session asked for strict structured output where the entry has the flag (D-205) and JSON mode with validation and one retry elsewhere.

## Source

> The AB3 brief from the design session, 2026-11-14: (4) a structured final answer for the Ask box, with strict structured output where the model entry has the flag and JSON mode with one retry where it has not, validated cards and suggestions, the UI drawn from the fields, and the bench scoring by the schema where it decides.