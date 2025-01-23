# Popeye error codes

## Severity list

- Severity 0: Ok
- Severity 1: Info
- Severity 2: Warning
- Severity 3: Error

## Container

| Error Code | Message                                                           | Severity | Info / Reference |
| ---------- | ----------------------------------------------------------------- | -------- | ---------------- |
| 100        | Untagged docker image in use                                      | 3        |                  |
| 101        | Image tagged "latest" in use                                      | 2        |                  |
| 102        | No probes defined                                                 | 2        |                  |
| 103        | No liveness probe                                                 | 2        |                  |
| 104        | No readiness probe                                                | 2        |                  |
| 105        | %s probe uses a port#, prefer a named port                        | 1        |                  |
| 106        | No resources requests/limits defined                              | 2        |                  |
| 107        | No resource limits defined                                        | 2        |                  |
| 108        | Unnamed port %d                                                   | 1        |                  |
| 109        | CPU Current/Request (%s/%s) reached user %d%% threshold (%d%%)    | 2        |                  |
| 110        | Memory Current/Request (%s/%s) reached user %d%% threshold (%d%%) | 2        |                  |
| 111        | CPU Current/Limit (%s/%s) reached user %d%% threshold (%d%%)      | 3        |                  |
| 112        | Memory Current/Limit (%s/%s) reached user %d%% threshold (%d%%)   | 3        |                  |
| 113        | Container image %s is not hosted on an allowed docker registry    | 3        |                  |

## Pod

| Error Code | Message                                              | Severity | Info / Reference |
| ---------- | ---------------------------------------------------- | -------- | ---------------- |
| 200        | Pod is terminating [%d/%d]                           | 2        |                  |
| 201        | Pod is terminating [%d/%d] %s                        | 2        |                  |
| 202        | Pod is waiting [%d/%d]                               | 3        |                  |
| 203        | Pod is waiting [%d/%d] %s                            | 3        |                  |
| 204        | Pod is not ready                                     | 3        |                  |
| 205        | Pod was restarted a number of times                  | 2        |                  |
| 206        | No PodDisruptionBudget defined                       | 1        |                  |
| 207        | Pod is in an unhappy phase                           | 3        |                  |
| 208        | Unmanaged pod detected. Best to use a controller     | 2        |                  |
| 209        | Pod is managed by multiple PodDisruptionBudgets (%s) | 2        |                  |

## Security

| Error Code | Message                                                              | Severity | Info / Reference |
| ---------- | -------------------------------------------------------------------- | -------- | ---------------- |
| 300        | Using "default" ServiceAccount                                       | 2        |                  |
| 301        | Connects to API Server? ServiceAccount token is mounted              | 2        |                  |
| 302        | Pod could be running as root user. Check SecurityContext/Image       | 2        |                  |
| 303        | Do you mean it? ServiceAccount is automounting APIServer credentials | 2        |                  |
| 304        | References a secret which does not exist                             | 3        |                  |
| 305        | References a docker-image "%s" pull secret which does not exist      | 3        |                  |
| 306        | Container could be running as root user. Check SecurityContext/Image | 2        |                  |
| 307        | %s references a non existing ServiceAccount: %q                      | 2        |                  |
| 308        | Uses "default" bound ServiceAccount. Could be a security risk        | 3        |                  |

## General

| Error Code | Message                                                     | Severity | Info / Reference |
| ---------- | ----------------------------------------------------------- | -------- | ---------------- |
| 400        | Used? Unable to locate resource reference                   | 1        |                  |
| 401        | Key "%s" used? Unable to locate key reference               | 1        |                  |
| 402        | No metrics-server detected                                  | 1        |                  |
| 403        | Deprecated %s API group "%s". Use "%s" instead              | 2        |                  |
| 404        | Deprecation check failed. %v                                | 1        |                  |
| 405        | Is this a jurassic cluster? Might want to upgrade K8s a bit | 2        |                  |
| 406        | K8s version OK                                              | 0        |                  |
| 407        | %s references %s %q which does not exist                    | 3        |                  |
| 666        | Lint internal error: %s                                     | 3        |                  |

## Workloads (Deployment and StatefulSet)

| Error Code | Message                                                                  | Severity | Info / Reference |
| ---------- | ------------------------------------------------------------------------ | -------- | ---------------- |
| 500        | Zero scale detected                                                      | 2        |                  |
| 501        | Unhealthy %d desired but have %d available                               | 3        |                  |
| 502        | MISSING                                                                  |          |                  |
| 503        | At current load, CPU under allocated. Current:%s vs Requested:%s (%s)    | 2        |                  |
| 504        | At current load, CPU over allocated. Current:%s vs Requested:%s (%s)     | 2        |                  |
| 505        | At current load, Memory under allocated. Current:%s vs Requested:%s (%s) | 2        |                  |
| 506        | At current load, Memory over allocated. Current:%s vs Requested:%s (%s)  | 2        |                  |
| 507        | Deployment references ServiceAccount %q which does not exist             | 3        |                  |
| 508        | No pods match controller selector: %s                                    | 3        |                  |

## HorizontalPodAutoscaler

| Error Code | Message                                                                       | Severity | Info / Reference |
| ---------- | ----------------------------------------------------------------------------- | -------- | ---------------- |
| 600        | HPA %s references a Deployment %s which does not exist                        | 3        |                  |
| 601        | HPA %s references a StatefulSet %s which does not exist                       | 3        |                  |
| 602        | Replicas (%d/%d) at burst will match/exceed cluster CPU(%s) capacity by %s    | 2        |                  |
| 603        | Replicas (%d/%d) at burst will match/exceed cluster memory(%s) capacity by %s | 2        |                  |
| 604        | If ALL HPAs triggered, %s will match/exceed cluster CPU(%s) capacity by %s    | 2        |                  |
| 605        | If ALL HPAs triggered, %s will match/exceed cluster memory(%s) capacity by %s | 2        |                  |

## Node

| Error Code | Message                                  | Severity | Info / Reference |
| ---------- | ---------------------------------------- | -------- | ---------------- |
| 700        | Found taint "%s" but no pod can tolerate | 2        |                  |
| 701        | Node is in an unknown condition          | 3        |                  |
| 702        | Node is not in ready state               | 3        |                  |
| 703        | Out of disk space                        | 3        |                  |
| 704        | Insufficient memory                      | 2        |                  |
| 705        | Insufficient disk space                  | 2        |                  |
| 706        | Insufficient PIDs on Node                | 3        |                  |
| 707        | No network configured on node            | 3        |                  |
| 708        | No node metrics available                | 1        |                  |
| 709        | CPU threshold (%d%%) reached %d%%        | 2        |                  |
| 710        | Memory threshold (%d%%) reached %d%%     | 2        |                  |
| 711        | Scheduling disabled                      | 2        |                  |
| 712        | Found only one master node               | 1        |                  |

## Namespace

| Error Code | Message               | Severity | Info / Reference |
| ---------- | --------------------- | -------- | ---------------- |
| 800        | Namespace is inactive | 3        |                  |

## PodDisruptionBudget

| Error Code | Message                                                                    | Severity | Info / Reference |
| ---------- | -------------------------------------------------------------------------- | -------- | ---------------- |
| 900        | Used? No pods match selector                                               | 2        |                  |
| 901        | MinAvailable (%d) is greater than the number of pods(%d) currently running | 2        |                  |

## PersistentVolume /PersistentVolumeClaim

| Error Code | Message                 | Severity | Info / Reference |
| ---------- | ----------------------- | -------- | ---------------- |
| 1000       | Available               | 1        |                  |
| 1001       | Pending volume detected | 3        |                  |
| 1002       | Lost volume detected    | 3        |                  |
| 1003       | Pending claim detected  | 3        |                  |
| 1004       | Lost claim detected     | 3        |                  |

## Service

| Error Code | Message                                                                   | Severity | Info / Reference |
| ---------- | ------------------------------------------------------------------------- | -------- | ---------------- |
| 1100       | No pods match service selector                                            | 3        |                  |
| 1101       | Skip ports check. No explicit ports detected on pod %s                    | 1        |                  |
| 1102       | Use of target port #%s for service port %s. Prefer named port             | 1        |                  |
| 1103       | Type Loadbalancer detected. Could be expensive                            | 1        |                  |
| 1104       | Do you mean it? Type NodePort detected                                    | 1        |                  |
| 1105       | No associated endpoints                                                   | 3        |                  |
| 1106       | No target ports match service port %s                                     | 3        |                  |
| 1107       | LoadBalancer detected but service sets externalTrafficPolicy to "Cluster" | 1        |                  |
| 1108       | NodePort detected but service sets externalTrafficPolicy to "Local"       | 1        |                  |
| 1109       | Only one Pod associated with this endpoint                                | 2        |                  |
| 1110       | Match EP has no subsets                                                   | 2        |                  |

## ReplicaSet

| Error Code | Message                                           | Severity | Info / Reference |
| ---------- | ------------------------------------------------- | -------- | ---------------- |
| 1120       | Unhealthy ReplicaSet %d desired but have %d ready | 3        |                  |

## NetworkPolicies

| Error Code | Message                                                    | Severity | Info / Reference |
| ---------- | ---------------------------------------------------------- | -------- | ---------------- |
| 1200       | No pods match %s pod selector                              | 2        |                  |
| 1201       | No namespaces match %s namespace selector                  | 2        |                  |
| 1202       | No pods match %s pod selector: %s                          | 2        |                  |
| 1203       | %s %s policy in effect                                     | 1        |                  |
| 1204       | Pod %s is not secured by a network policy                  | 2        |                  |
| 1205       | Pod ingress and egress are not secured by a network policy | 2        |                  |
| 1206       | No pods matched %s IPBlock %s                              | 2        |                  |
| 1207       | No pods matched except %s IPBlock %s                       | 2        |                  |
| 1208       | No pods match %s pod selector: %s in namespace: %s         | 2        |                  |

## RBAC

| Error Code | Message                                   | Severity | Info / Reference |
| ---------- | ----------------------------------------- | -------- | ---------------- |
| 1300       | References a %s (%s) which does not exist | 2        |                  |

## Ingress

| Error Code | Message                                                       | Severity | Info / Reference |
| ---------- | ------------------------------------------------------------- | -------- | ---------------- |
| 1400       | Ingress LoadBalancer port reported an error: %s               | 3        |                  |
| 1401       | Ingress references a service backend which does not exist: %s | 3        |                  |
| 1402      | Ingress references a service port which is not defined: %s     | 3        |                  |
| 1403      | Ingress backend uses a port#, prefer a named port: %d          | 1        |                  |
| 1404      | Invalid Ingress backend spec. Must use port name or number     | 3        |                  |

## CronJob

| Error Code | Message                                   | Severity | Info / Reference |
| ---------- | ----------------------------------------- | -------- | ---------------- |
| 1500       | %s is suspended                           | 2        |                  |
| 1501       | No active jobs detected                   | 1        |                  |
| 1502       | CronJob has not run yet or is failing     | 2        |                  |
| 1503       | Warning found: %s                         | 2        |                  |

## CiliumIdentity

| Error Code | Message                                                     | Severity | Info / Reference |
| ---------- | ----------------------------------------------------------- | -------- | ---------------- |
| 1600       | Stale? unable to locate matching Cilium Endpoint            | 2        |                  |
| 1601       | Unable to assert namespace label: %q                        | 2        |                  |
| 1602       | References namespace which does not exists: %q              | 2        |                  |
| 1603       | Missing security namespace label: %q                        | 2        |                  |
| 1604       | Namespace mismatch with security labels namespace: %q vs %q | 2        |                  |

## CiliumEndpoint

| Error Code | Message                                      | Severity | Info / Reference |
| ---------- | -------------------------------------------- | -------- | ---------------- |
| 1700       | No cilium endpoints matched %s selector      | 3        |                  |
| 1701       | No nodes matched node selector               | 3        |                  |
| 1702       | References an unknown node IP: %q            | 3        |                  |
| 1703       | Pod owner is not in a running state: %s (%s) | 3        |                  |
| 1704       | References an unknown owner ref: %q          | 3        |                  |
