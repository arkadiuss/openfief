---
title: A keyless trusted signer that pins a source repository or source ref must count a signature only when the certificate's Fulcio source repository URI and source ref extensions equal them, OpenFief's CI signer must pin https://github.com/arkadiuss/openfief at refs/heads/main, and a signer of kind OpenFiefCI without both must be refused.
status: open
text_ref: docs/DESIGN.md#D-181
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-181
  - lane:strict
semantic_text: A keyless trusted signer that pins a source repository or source ref must count a signature only when the certificate's Fulcio source repository URI and source ref extensions equal them, OpenFief's CI signer must pin https://github.com/arkadiuss/openfief at refs/heads/main, and a signer of kind OpenFiefCI without both must be refused.
logic_claims:
  - CLAIM-4417F27FA2EDB08C
semantic_clauses:
  - A keyless trusted signer that pins a source repository or source ref must count a signature only when the certificate's Fulcio source repository URI and source ref extensions equal them, OpenFief's CI signer must pin https://github.com/arkadiuss/openfief at refs/heads/main, and a signer of kind OpenFiefCI without both must be refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7053c5b4e4c27ad36a7b70c4e52f956b7803f2b7078b308996a76e0de19c5d46
semantic_inventory:
  - claim_key: CLAIM-4417F27FA2EDB08C
    claim_text: A keyless trusted signer that pins a source repository or source ref must count a signature only when the certificate's Fulcio source repository URI and source ref extensions equal them, OpenFief's CI signer must pin https://github.com/arkadiuss/openfief at refs/heads/main, and a signer of kind OpenFiefCI without both must be refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 335
origin:
  kind: agent
  recorded_at: '2026-11-02T20:00:19.214Z'
id: REQ-plugins-signer-source-pinned
type: req
---
A keyless trusted signer that pins a source repository or source ref must count a signature only when the certificate's Fulcio source repository URI and source ref extensions equal them, OpenFief's CI signer must pin https://github.com/arkadiuss/openfief at refs/heads/main, and a signer of kind OpenFiefCI without both must be refused.

## Context

The security review found that OpenFief's CI identity was matched by issuer and subject alone. The subject is the reusable images.yml workflow at refs/heads/main in a public repository, and for a reusable workflow Fulcio certifies the called workflow as the subject, so any repository can call it with an OIDC token, build its own code in that job and get a certificate the CLI and the operator accepted as openfief-ci. Fulcio's source repository URI and ref extensions name the run's own repository and ref, the caller's, which no other repository can forge. The author decided to pin OpenFief's CI to the source repository and ref (D-181). Owner and catalog signers may pin them too, which a reusable signing workflow needs; documents without them read as before. Refines REQ-plugins-signer-identity-match and REQ-plugins-trusted-signers-only (D-45, D-56, GR-12). Modeled by Claude Code for the P0 security fixes before the alpha release, from the security review's finding that OpenFief's CI identity could be obtained by any repository; the author decided the fix as D-181.

## Source

> D-181: the CI signer is pinned to the source repository and ref, arkadiuss/openfief at refs/heads/main.

docs/DESIGN.md D-181 (author decision on the security review's finding P0-1), D-45, D-56, GR-12
