---
title: The CA bundle is mounted only when expected, its state and data checksum are pure, and a moved host waits
status: active
sourceFile: internal/plugins/ca_bundle_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:55.382Z'
id: TEST-plugins-ca-bundle
type: test
---
Table tests over the plugins domain: TestEffectiveCABundle compiles a synthetic connector whose manifest asks for a token and the optional ca-bundle role and checks that the bundle is mounted only when the instance expects one and the grant approves it, while both stay requested; TestSetExtraCA and TestPlanUpgradeKeepsExtraCA check the manifest gate and that an upgrade keeps the setting only while the role stays; TestCABundleStateOf covers every state; TestSecretDataChecksum checks order independence, the empty digest and that moving bytes between a key and a value changes the checksum; TestConfigHostChangeKeepsApprovedAllowList plans a rollout with a base URL moved to another host and checks the old approved egress stays enforced, the new host is pending and the upgrade diff shows both. A false pass would need Effective and Requested to agree by accident on every row.