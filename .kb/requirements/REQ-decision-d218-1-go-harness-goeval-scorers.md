---
title: The evals run in a Go harness in OpenFief whose scorers implement the interface of goeval, scorers that know OpenFief's data stay in OpenFief, and general ones go into goeval.
status: open
origin:
  kind: import
  ref: design-decisions:D-218#1
  recorded_at: '2026-11-14T08:23:24.899Z'
firmness: firm
semantic_text: The evals run in a Go harness in OpenFief whose scorers implement the interface of goeval, scorers that know OpenFief's data stay in OpenFief, and general ones go into goeval.
text_ref: design-decisions:D-218#1
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 40b0233c0e519e7fb69ed061c962dc0b71e96a75e22532ce825e114e799ed480
semantic_inventory:
  - claim_key: CLAIM-7675BDDE6EE8EF47
    claim_text: The evals run in a Go harness in OpenFief whose scorers implement the interface of goeval, scorers that know OpenFief's data stay in OpenFief, and general ones go into goeval
    role: descriptive
    status: missing
    span:
      start: 0
      end: 174
logic_claims:
  - CLAIM-7675BDDE6EE8EF47
id: REQ-decision-d218-1-go-harness-goeval-scorers
type: req
---
The evals run in a Go harness in OpenFief whose scorers implement the interface of goeval, scorers that know OpenFief's data stay in OpenFief, and general ones go into goeval.

## Context

Author decision D-218 (the harness of the evals).

## Source

> **[D-218]** D-217's evals run in a Go harness in OpenFief whose scorers implement the interface of goeval (github.com/datar-psa/goeval). Scorers that know OpenFief's data, such as numbers backed by the cards shown, stay in OpenFief; general ones, such as tool-call correctness and a judge on an OpenAI-compatible endpoint, go into goeval.
