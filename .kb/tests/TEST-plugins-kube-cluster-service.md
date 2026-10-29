---
title: The CLI's cluster adapter checks the in-cluster service, grants the operator read access to it and builds the ingress rule
status: active
sourceFile: internal/plugins/kube/clusterservice_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:03.653Z'
id: TEST-plugins-kube-cluster-service
type: test
---
Over the controller-runtime fake client: Resources carries spec.clusterService, egressFromConfig and the grant's clusterServices; ResourcesYAML adds the Role and RoleBinding; ServiceAccess lets only the operator get only that Service; ClusterService reads the target port and selector and sees a namespace's NetworkPolicies, refusing a missing port or Service; IngressRule admits the instance's pods; install creates and uninstall deletes the access.