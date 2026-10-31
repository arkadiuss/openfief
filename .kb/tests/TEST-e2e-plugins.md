---
title: The operator deploys a signed first-party plugin on k3d and kind, isolated as granted, and removes it again
status: active
sourceFile: test/e2e/plugins_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:38.282Z'
id: TEST-e2e-plugins
type: test
verification_scope: end_to_end
---
Runs on a local k3d or kind cluster after task e2e:images pushed the nbp-fx image built from this checkout to a registry container on the cluster's Docker network and TestMain signed it there with an ECDSA key made in memory for the run, added as an owner's trusted signer in ConfigMap openfief-trusted-signers before the operator started. TestOperatorDeploysSignedPlugin applies a PluginGrant and a Plugin nbp-e2e with the image's own manifest and a config whose start lies in 2999, so the connector never calls NBP. It waits for Ready, failing at once on a refused signature or a stuck pod; checks SignatureVerified is SignedByTrustedSigner, status.signature names the owner's signer as unverified, and status.effective holds only the requested and granted host although the grant also names example.org; compares the running pod's ServiceAccount, automount, security context, image and projected token audiences with the expected ones; runs core connector sync inside the core pod until it reports an empty batch; checks that the client and outsider pods get no connection to the plugin's Service; checks the allow-list entry is exactly api.nbp.pl:443; then deletes the Plugin and waits for its Deployment, Service, NetworkPolicy and allow-list entry to go. A false pass would need the operator to run an unsigned image, which the signature checks rule out, or the probes to miss a listener, which the core's successful sync rules out.