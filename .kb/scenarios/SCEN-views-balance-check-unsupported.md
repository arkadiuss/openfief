---
title: Open XTB CFDs are listed as flagged and unsupported, not checked
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:12.047Z'
id: SCEN-views-balance-check-unsupported
type: scenario
---
Given an XTB instance whose latest completed upload left open CFD positions out as unsupported rows, when the owner's balance check runs, then the report lists that instance with its latest batch and the number of rows left out, flagged and not checked; when a later batch of the instance leaves none out, the instance is no longer listed.