---
title: A sleeve without exactly one known role is rejected
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:13:23.655Z'
id: SCEN-portfolio-sleeve-role
type: scenario
---
Given a sleeve with an ID and a name, when its role is defensive or offensive, then it is valid; when its role is missing or not one of those two, then validation fails with an invalid sleeve error and the sleeve is not usable. Assumes roles are stored as the words defensive and offensive and parsed case-sensitively.