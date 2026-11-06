---
title: On k3d and kind, a connector's credentials stay in its Secret, mounted only into its pod, and uninstalling it keeps its transactions marked as from a removed instance
status: active
sourceFile: test/e2e/connector_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-06T19:43:39.648Z'
id: TEST-e2e-connector-credentials
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-connector-test-connector-credentials-and-uninstall
      target: default
  success_policy: all_required_first_attempt
---
TestConnectorCredentialsAndUninstall creates the Secret e2e-books--credentials with a random synthetic credential, binds the synthetic connector instance to the demo XTB account and applies its Plugin and a grant naming the Secret. It checks that only the instance pods use the Secret, syncs the instance so that the connector reads the credential and books one deposit, then searches every ConfigMap, every pod spec, the Plugins, the PluginGrants and a pg_dumpall of the database for the credential. openfief plugin uninstall then has to remove the Plugin, grant, Secret, Deployment, Service, NetworkPolicy, ServiceAccount and allow-list entry, and the read API has to keep listing the deposit with connector_instance_removed set. The credential is never logged.