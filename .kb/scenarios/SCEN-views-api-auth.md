---
title: A call without a valid bearer token gets no data, over any protocol
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:43:19.364Z'
id: SCEN-views-api-auth
type: scenario
---
Given the read API with a token verifier, when a REST GET or a Connect call carries no Authorization header, another scheme, an empty bearer token or a token the verifier rejects, then it is refused as unauthenticated (401 with a WWW-Authenticate Bearer challenge over REST) without naming any account; when the verifier cannot check the token, then it fails as internal (500); only a token the verifier accepts reaches the data.