---
title: TokenReview checks the openfief-core audience and the allow-list ConfigMap is followed fail-closed
status: active
sourceFile: internal/plugins/kube/kube_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:52.741Z'
id: TEST-plugins-kube
type: test
---
Uses the client-go fake clientset. TokenReviewer tests answer TokenReview creates through a reactor and check that the review asks for the openfief-core audience, returns the username when authenticated for it, and refuses an unauthenticated token, a token authenticated only for another audience and an API error, never echoing the token. Allow-list tests start a watch on a synthetic openfief-egress-allowlist ConfigMap, then update it, make it undecodable, fix it and delete it, waiting with a deadline for each state: updates apply, an undecodable or deleted ConfigMap leaves no plugin with a list and is reported, another ConfigMap is ignored, a missing ConfigMap means no lists, and a cancelled context returns ErrNotSynced. A false pass would be keeping the last good list after a bad update; that state is asserted.