---
title: Outside a cluster no caller reaches the AuthService, and the UI presents only its projected token
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T10:54:14.534Z'
id: SCEN-identity-no-auth-bypass
type: scenario
---
Given a core started outside a Kubernetes cluster, when any caller presents any token, even one that a fake cluster's TokenReview would accept, then the AuthService refuses it as not a view plugin. Given a core in a fake cluster whose TokenReview accepts a synthetic view plugin token, when the UI's client calls the AuthService, then it is admitted through the same TokenReview check as in a real cluster. Given the UI plugin, when it calls the core, then it presents the token read from its projected token file and nothing else.