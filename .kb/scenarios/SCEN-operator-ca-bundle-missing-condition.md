---
title: An expected CA bundle without a stamp is reported from the spec
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:59.871Z'
id: SCEN-operator-ca-bundle-missing-condition
type: scenario
---
Given a Plugin whose spec sets extraCa, when no spec.caBundleSha256 is stamped then its status has CABundleMissing True with reason ChecksumNotStamped and a message telling the owner to run openfief plugin ca-bundle set, and the Deployment is applied anyway; when a checksum is stamped the condition is False with reason ChecksumStamped; when extraCa is not set the condition is absent. The operator decides from the spec alone and never reads the Secret, so a Secret created by hand without a stamp changes nothing.