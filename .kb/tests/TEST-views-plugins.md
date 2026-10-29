---
title: A Plugin's status maps to one trust level, and a transaction's instance is marked removed only when known gone
status: active
sourceFile: internal/views/plugins_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:20.213Z'
id: TEST-views-plugins
type: test
---
Table tests of views.PluginOf over synthetic Plugin records: SignatureVerified True with status.signature Verified is verified, Unverified by OpenFiefCI, Publisher or Owner maps to that signer kind, True without a signature field is signed by a trusted signer, False is refused, Unknown or absent is not checked; the image digest, phase (Ready's reason, Removing while deleted, Pending without a condition) and signer shown only for a trusted signature. TestConnectorInstance reads ingest.ConnectorKeySource's own output back, so the two can't drift. TestOrigins marks only a plugin:<instance> source whose instance is not installed, and nothing while the installed plugins are unknown.