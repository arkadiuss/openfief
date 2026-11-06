---
title: A plugin image named by tag is refused, by the API server or by the operator
status: active
tags:
  - acceptance:p10-11
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:37.042Z'
id: SCEN-e2e-plugin-tag-refused
type: scenario
---
Given a grant that approves everything the nbp-fx manifest requests, when an owner applies a Plugin whose image is a tag alone, then the API server refuses it as invalid and no Plugin exists; when the image names a tag next to its digest, then the operator reports SignatureVerified False with reason InvalidImage and Ready False with reason InvalidSpec, and runs nothing. Runs on both kind and k3d.