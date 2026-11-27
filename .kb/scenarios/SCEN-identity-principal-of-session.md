---
title: A viewer's session is never the owner, whichever way its principal is built
status: active
origin:
  kind: agent
  ref: the owner's items for alpha.2, from the reviews of main 0bfeb62
  recorded_at: '2026-11-27T09:04:48.400Z'
id: SCEN-identity-principal-of-session
type: scenario
---
Given the owner's session, a viewer's read session, a viewer's session with owner scope, a viewer's session from the demo authenticator and a demo session, the principal built from each is the owner, the viewer, the viewer, the viewer and the demo session; a principal built without saying whose session it is is nobody, which the authorizer refuses; the maintainer's note is not shown to a viewer granted fewer than every account, whatever the household holds for the owner.