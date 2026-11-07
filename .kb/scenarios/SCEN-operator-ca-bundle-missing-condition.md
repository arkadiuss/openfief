---
title: A missing expected CA bundle is reported, not waited for
status: active
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:59.871Z'
id: SCEN-operator-ca-bundle-missing-condition
type: scenario
---
Given a running Plugin that expects an extra CA bundle, when its Secret <instance>--ca-bundle does not exist then the Plugin status has CABundleMissing True with reason SecretMissing and a message telling the owner to create it with ca.crt and the instance label, and the Deployment is applied anyway; when it exists the condition is False SecretPresent; when the operator may not read it the condition is Unknown SecretUnreadable naming the Role to apply; when no bundle is expected the condition is absent. No message holds the Secret data.