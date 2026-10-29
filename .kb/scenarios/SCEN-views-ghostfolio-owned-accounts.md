---
title: The view writes only into the Ghostfolio accounts it derived and created
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:22.412Z'
id: SCEN-views-ghostfolio-owned-accounts
type: scenario
---
Given a record with accounts acct-bonds and acct-cash and a Ghostfolio user who has an account of their own, when the view instance named ghostfolio syncs, then it creates one Ghostfolio account per record account with the UUID derived from the instance name and the account ID, named after the account, in the user's base currency; when it syncs again, then it finds the same accounts and creates none; and the user's own account and its activities are never changed.