---
title: A plugin instance can never mount or delete one of the platform's Secrets
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T22:19:40.715Z'
id: SCEN-plugins-secrets-never-platform
type: scenario
---
Given the platform's own objects in the openfief namespace, all named openfief or openfief-*, such as the database credentials openfief-db-app, when an owner installs an image whose repository is named openfief, or names an instance openfief or openfief-db, or a grant names a Secret openfief-db-app, then the install and the operator refuse it with ErrReservedName; and when an instance is uninstalled, only the Secrets labelled openfief.io/instance with its name are deleted, each read again just before the delete, while unlabelled ones and another instance's are left and listed.