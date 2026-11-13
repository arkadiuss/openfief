---
title: The release bundle's credentials Job runs the core's pinned image, and its Role is the only one in the install that may create a Secret
status: active
origin:
  kind: agent
  recorded_at: '2026-11-13T07:41:54.582Z'
id: SCEN-deploy-credentials-job-core-image-create-only
type: scenario
---
Given the release bundle rendered from the real deploy/, then the Job labelled app.kubernetes.io/name=openfief-init-credentials, named openfief-init-credentials-<the bundle's version, lowercased, dots as dashes> (a Job's pod template is immutable and the image changes each release, so each release's Job has a name of its own, and the base's name is gone), runs the core's image at the digest the core runs, with the argument init-credentials, and no other third-party image is named for it; the Role openfief-init-credentials grants create on secrets and nothing else; it is bound to the Job's ServiceAccount alone; and no other Role or ClusterRole in the install, wildcards included, grants create on secrets.