---
title: Verified, unverified or refused
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:50:45.564Z'
id: SCEN-plugins-trust-decision
type: scenario
---
Given a catalog that pins digest A to a publisher, when digest A carries the publisher's signature then it is Verified; when digest B carries the publisher's or OpenFief CI's or an owner signer's signature then it is Unverified; and when an image carries no signature or only signatures by unknown signers then it is refused and never deployed.