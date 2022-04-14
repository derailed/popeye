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
| 106        | No resources requests/limits defined                              | 2        |
| 107        | No resource limits defined                                        | 2        |                  |
| 108        | Unnamed port %d                                                   | 1        |                  |
| 109        | CPU Current/Request (%s/%s) reached user %d%% threshold (%d%%)    | 2        |                  |
| 110        | Memory Current/Request (%s/%s) reached user %d%% threshold (%d%%) | 2        |                  |
| 111        | CPU Current/Limit (%s/%s) reached user %d%% threshold (%d%%)      | 3        |                  |
| 112        | Memory Current/Limit (%s/%s) reached user %d%% threshold (%d%%)   | 3        |                  |

## Pod

| Error Code | Message                                          | Severity | Info / Reference |
| ---------- | ------------------------------------------------ | -------- | ---------------- |
| 200        | Pod is terminating [%d/%d]                       | 2        |                  |
| 201        | Pod is terminating [%d/%d] %s                    | 2        |                  |
| 202        | Pod is waiting [%d/%d]                           | 3        |                  |
| 203        | Pod is waiting [%d/%d] %s                        | 3        |                  |
| 204        | Pod is not ready                                 | 3        |                  |
| 205        | Pod was restarted a number of times              | 2        |                  |
| 206        | No PodDisruptionBudget defined                   | 1        |                  |
| 207        | Pod is in an unhappy phase                       | 3        |                  |
| 208        | Unmanaged pod detected. Best to use a controller | 2        |                  |

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

## General

| Error Code | Message                                                     | Severity | Info / Reference |
| ---------- | ----------------------------------------------------------- | -------- | ---------------- |
| 400        | Used? Unable to locate resource reference                   | 1        |                  |
| 401        | Key "%s" used? Unable to locate key reference               | 1        |                  |
| 402        | No metric-server detected %v                                | 1        |                  |
| 403        | Deprecated %s API group "%s". Use "%s" instead              | 2        |                  |
| 404        | Deprecation check failed. %v                                | 1        |                  |
| 405        | Is this a jurassic cluster? Might want to upgrade K8s a bit | 2        |                  |
| 406        | K8s version OK                                              | 0        |                  |

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

## ReplicaSet

| Error Code | Message                                           | Severity | Info / Reference |
| ---------- | ------------------------------------------------- | -------- | ---------------- |
| 1120       | Unhealthy ReplicaSet %d desired but have %d ready | 3        |                  |

## NetworkPolicies

| Error Code | Message                                   | Severity | Info / Reference |
| ---------- | ----------------------------------------- | -------- | ---------------- |
| 1200       | No pods match %s pod selector             | 2        |                  |
| 1201       | No namespaces match %s namespace selector | 2        |                  |

## RBAC

| Error Code | Message                                   | Severity | Info / Reference |
| ---------- | ----------------------------------------- | -------- | ---------------- |
| 1300       | References a %s (%s) which does not exist | 2        |                  |
