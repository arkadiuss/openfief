---
title: An enrollment link creates the first passkey and signs the owner in
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:50.316Z'
id: SCEN-ui-enrollment
type: scenario
---
Given an enrollment link with its code in the fragment, when the page posts the code to /enroll/begin and the new credential to /enroll/finish, then BeginRegistration gets the code and no session, and the session that FinishRegistration returns goes in the cookie; with an owner session and no code, the session authorizes the registration instead.