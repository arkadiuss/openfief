---
title: validate.sh builds every demo layer and rejects unpinned images, unverified OCI sources or unknown variables, and keeps the demo's model layer, which the live demo includes, consistent
status: active
sourceFile: scripts/tests/demo-validate.test.sh
origin:
  kind: agent
  ref: 'Demo operations brief from the design session: ship the demo''s telemetry to Grafana Cloud with no visitor personal data, 2026-11-19'
  recorded_at: '2026-11-19T16:24:01.184Z'
id: TEST-tooling-demo-validate
type: test
verification_scope: integration
---
Runs scripts/demo/validate.sh with the pinned kustomize on copies of deploy/ in a throwaway root. The committed manifests pass; an OCI source whose verification provider is not cosign fails naming it; removing the images list of deploy/demo/openfief (images by tag only) fails naming the image; a git-sourced Flux Kustomization path that does not exist fails naming the path; and a substituted layer using a variable the demo settings do not define fails naming it. The model layer (deploy/demo/model, included by deploy/demo/openfief): a route to a model its provider instance does not list, a route for the maintainer, the layer left out of the live demo (its spend guards then missing), a layer that does not build and the placeholder Turnstile site key each fail naming the problem, as does each of D-208's guards changed. Every case also checks the message, so a failure for the wrong reason does not pass. A PluginGrant that leaves out a Secret its Plugin's manifest asks for (the openrouter grant without openrouter--ca-bundle) fails naming the grant and the Secret. The telemetry layer (deploy/demo/telemetry): the layer left out of the Flux stack, the scrub missing from a pipeline or behind another processor, a second exporter, a scrub that fails open, no statement for a context, the log body or a span's name and status message, a key or pattern left out (Authorization, question and prompt, the token prefixes), a literal endpoint, token or Grafana address, a Secret in git, egress to another port or without the private ranges left out, an ingress port that is not 4318, a writable log mount, another host path, a credential the pod does not wait for, kept capabilities, an unpinned image and the core sending anywhere but the collector's Service each fail naming the problem.