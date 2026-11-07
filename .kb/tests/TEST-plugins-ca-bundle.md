---
title: CA bundle files are checked and hashed, the state comes from the spec, and the bundle mounts only when expected
status: active
sourceFile: internal/plugins/ca_bundle_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:55.382Z'
id: TEST-plugins-ca-bundle
type: test
---
Table tests over the plugins domain: TestParseCABundle feeds one and two self-signed certificates made for the test, an empty file, whitespace, text, trailing text, a private key, a key after a certificate and a CERTIFICATE block that does not parse, checking the certificate count, the checksum and each refusal error; TestCABundleChecksum pins the SHA-256 of known inputs, as sha256sum prints them; TestCABundleStateOf covers every combination of extraCa and a stamp; TestPluginStampCABundle checks stamping needs an expected bundle and a hex SHA-256 and that an empty one clears it; TestEffectiveCABundle, TestSetExtraCA and TestPlanUpgradeKeepsExtraCA check the bundle is mounted only when expected and approved, that turning it off clears the stamp, that Validate refuses a bad stamp, and that an upgrade keeps setting and stamp only while the role stays; TestConfigHostChangeKeepsApprovedAllowList plans a rollout with a base URL moved to another host and checks the old approved egress stays enforced, the new host pending and the upgrade diff shows both.