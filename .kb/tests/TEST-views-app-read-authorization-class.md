---
title: Every read procedure of Views asks the authorizer about an account or is on the list of reads that name none
status: active
sourceFile: internal/views/app/readclass_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the owner's items for alpha.2, from the reviews of main 0bfeb62
  recorded_at: '2026-11-27T08:30:56.184Z'
id: TEST-views-app-read-authorization-class
type: test
---
TestEveryReadProcedureAsksTheAuthorizerOrIsHouseholdLevel parses the views/app package, takes every exported method of Views whose first parameter is a context as a read procedure, follows its calls through the package's own functions, and fails for one that reaches no call of MayAct or ActableAccounts and is not in unauthorizedByAccount, and for a listed name that is no read procedure.