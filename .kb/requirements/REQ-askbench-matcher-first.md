---
title: The Ask box bench must run the box's own command matcher before the model, so that a question the box answers itself is tagged matcher in the report, is never asked of the model and counts toward neither its safety nor its quality, as in production, where it never reaches the model.
status: open
firmness: default
required_evidence: unverified
semantic_text: The Ask box bench must run the box's own command matcher before the model, so that a question the box answers itself is tagged matcher in the report, is never asked of the model and counts toward neither its safety nor its quality, as in production, where it never reaches the model.
tags:
  - askbench
origin:
  kind: agent
  ref: 'AB5 brief from the design session: bench fixes before the next run and the sync card''s error class (D-194, D-198, D-203, D-217), 2026-11-14'
  recorded_at: '2026-11-14T18:29:32.000Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 72785eb9a2562f3a98714cf842d1c429f65503f42bb310fc20575575809666ee
semantic_inventory:
  - claim_key: CLAIM-22079A5D01309A14
    claim_text: The Ask box bench must run the box's own command matcher before the model, so that a question the box answers itself is tagged matcher in the report, is never asked of the model and counts toward neither its safety nor its quality, as in production, where it never reaches the model
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 282
    subjects:
      - ask_bench.matcher_first
logic_claims:
  - CLAIM-22079A5D01309A14
id: REQ-askbench-matcher-first
type: req
pair_reviews:
  - with: REQ-ui-ask-model-fixed-commands-first
    verdict: no_conflict
    separated_by: the bench's report tagging and scoring of a matched question, which the bench rule states, against the UI's routing of a fixed command at run time, which the UI rule states; the bench runs the UI's own matcher, so neither rule changes what the other does
    claims:
      - claim_key: CLAIM-22079A5D01309A14
    origin:
      kind: agent
      ref: Author's decision to upgrade to kibi 2.45 and review the subjects, 2026-11-20
---
The Ask box bench must run the box's own command matcher before the model, so that a question the box answers itself is tagged matcher in the report, is never asked of the model and counts toward neither its safety nor its quality, as in production, where it never reaches the model.

## Context

D-203 has the box answer its fixed commands itself and send the model anything else, so a question the matcher answers never reaches a model in production. The bench sent such questions to the model all the same, which graded the model on questions it never sees and, for the injection cases, tested a payload the page would have shown without any model. The design session decided the bench runs the real matcher first and the cases are worded to reach the model (2026-11-14).

## Source

> AB5 brief, decision 4: MATCHER FIRST: D-203, the box answers its fixed commands itself, anything else goes to the model. Make the bench faithful to production: run the real matcher first; a case the matcher answers is tagged matcher in the report and does not count toward the model's quality.
