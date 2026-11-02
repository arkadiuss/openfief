---
title: Owner and catalog signers may pin a source repository and ref
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T20:00:07.007Z'
id: SCEN-plugins-signer-source-documents
type: scenario
---
Given an owner's signers.json or a catalog entry whose keyless signer sets sourceRepository and sourceRef, when the document is read then the signer keeps both and matches only signatures from that repository and ref; when the source repository is not an https URL or the source ref is not a refs/ ref, the whole document is refused; a document without the fields reads as before, and a signer of kind OpenFiefCI without both is refused.