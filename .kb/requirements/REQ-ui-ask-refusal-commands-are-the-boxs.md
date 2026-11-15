---
title: A refusing answer of the Ask box carries the box's own commands as suggestions, the same for every refusal and in English, the only language its matcher understands, and anything a model writes in suggestions is ignored, so a slip in it never fails a correct refusal or forces another round.
status: open
tags:
  - askbench
origin:
  kind: agent
  ref: 'Design session decision: the box owns a refusal''s command suggestions and vets clarify candidates (D-194), 2026-11-15'
  recorded_at: '2026-11-15T10:32:45.812Z'
semantic_text: A refusing answer of the Ask box carries the box's own commands as suggestions, the same for every refusal and in English, the only language its matcher understands, and anything a model writes in suggestions is ignored, so a slip in it never fails a correct refusal or forces another round.
logic_claims:
  - CLAIM-C31B2503BB996A06
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: default
required_evidence: unverified
semantic_source_hash: 483a7ba278e7f444dc87ee8d7859a10ab3f96e9cafa2a3fa54d54ed1a7234cf1
semantic_inventory:
  - claim_key: CLAIM-C31B2503BB996A06
    claim_text: A refusing answer of the Ask box carries the box's own commands as suggestions, the same for every refusal and in English, the only language its matcher understands, and anything a model writes in suggestions is ignored, so a slip in it never fails a correct refusal or forces another round
    role: descriptive
    status: ontology_gap
    span:
      start: 0
      end: 290
id: REQ-ui-ask-refusal-commands-are-the-boxs
type: req
---
A refusing answer of the Ask box carries the box's own commands as suggestions, the same for every refusal and in English, the only language its matcher understands, and anything a model writes in suggestions is ignored, so a slip in it never fails a correct refusal or forces another round.

## Context

A traced bench run showed a correct refusal failing on a malformed suggestions array and the final round then turning it into a greeting. The box answers what it doesn't understand with the commands it does (D-194), so it, not the model, attaches them.

## Source

Design session decision, 2026-11-15: on kind refuse the box appends its own command suggestions from the matcher's command list; the schema drops suggestions; the bench grades the box's.
