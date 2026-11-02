---
title: A denied change is refused before anything happens
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T22:46:18.121Z'
id: SCEN-core-authz-change-refused
type: scenario
---
Given an owner session and an authorizer that denies a change of one account, when an owner service is asked to update, bind, delete or remap that account, enter or delete a bond lot in it, upload a statement for it, or decide a transfer with a leg in it, then it fails with ErrNotOwner and stores, binds and hands over nothing.