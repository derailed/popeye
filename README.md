<img src="assets/popeye.png" align="right" width="250" heigh="auto">

# Popeye - A Kubernetes Cluster Sanitizer

Popeye is a utility that cruises Kubernetes cluster resources and reports potential
issues with your deployment manifests and configurations. By scanning your
clusters, it detects misconfigurations and ensure best practices are in place thus
preventing potential future headaches. It aims at reducing the cognitive *over*load
one faces when managing and operating a Kubernetes cluster in the wild. Popeye
is a readonly tool, it does not change or update any of your Kubernetes resources or
configurations in any way!

<br/>
<br/>

---

[![Go Report Card](https://goreportcard.com/badge/github.com/derailed/popeye?)](https://goreportcard.com/report/github.com/derailed/popeye)
[![Build Status](https://travis-ci.com/derailed/popeye.svg?branch=master)](https://travis-ci.com/derailed/popeye)
[![release](https://img.shields.io/github/release-pre/derailed/popeye.svg)](https://github.com/derailed/popeye/releases)
[![Get it from the Snap Store](https://snapcraft.io/static/images/badges/en/snap-store-black.svg)](https://snapcraft.io/popeye)

---

## Installation

Popeye is available on Linux, OSX and Windows platforms.

* Binaries for Linux, Windows and Mac are available as tarballs in
  the [release](https://github.com/derailed/popeye/releases) page or
  via the SnapCraft link above.

* For OSX using Homebrew

   ```shell
   brew tap derailed/popeye && brew install popeye
   ```

* Building from source
   Popeye was built using go 1.12+. In order to build Popeye from source you must:
   1. Clone the repo
   2. Set env var *GO111MODULE=on*
   3. Add the following command in your go.mod file

      ```text
      replace (
        github.com/derailed/popeye => MY_POPEYE_CLONED_GIT_REPO
      )
      ```

   4. Build and run the executable

        ```shell
        go run main.go
        ```

## Sanitizers

Popeye scans your cluster for best practices and potential issues.
Currently, Popeye only looks at nodes, namespaces, pods and services.
More will come soon! We are hoping Kubernetes friends will pitch'in
to make Popeye even better.

The aim of the sanitizers is to pick up on misconfigurations ie things
like ports mismatch, dead or unused resources, metrics utilization,
probes, container images, RBAC rules, naked resources, etc...

Here is a list of sanitizers in place for the current release.

| Resource       | Sanitizers                                                              |
|----------------|-------------------------------------------------------------------------|
| Node           |                                                                         |
|                | Conditions ie not ready, out of mem/disk, network, pids, etc            |
|                | Pod tolerations referencing node taints                                 |
|                | CPU/MEM utilization metrics, trips if over limits (default 80% CPU/MEM) |
| Namespace      |                                                                         |
|                | Inactive                                                                |
|                | Dead namespaces                                                         |
| Pod            |                                                                         |
|                | Pod status                                                              |
|                | Containers statuses                                                     |
|                | ServiceAccount presence                                                 |
|                | CPU/MEM on containers over a set CPU/MEM limit (default 80% CPU/MEM)    |
|                | Container image with no tags                                            |
|                | Container image using `latest` tag                                      |
|                | Resources request/limits presence                                       |
|                | Probes liveness/readiness presence                                      |
|                | Named ports and their references                                        |
| Service        |                                                                         |
|                | Endpoints presence                                                      |
|                | Matching pods labels                                                    |
|                | Named ports and their references                                        |
| ServiceAccount |                                                                         |
|                | Dead SA ie used by CRB/RB but no matching pod ServiceAccount reference  |

## The Command Line

You can use Popeye standalone or using a spinach yaml config to
tune the sanitizer. Details about the Popeye configuration file are below.

```shell
# Dump version info
popeye version
# Popeye a cluster using your current kubeconfig environment.
popeye
# Popeye uses a spinach config file of course! aka spinachyaml!
popeye -f spinach.yml
# Popeye a cluster using a kubeconfig context.
popeye --context olive
# Stuck?
popeye help
```

## Screenshots

### Cluster D Score

<img src="assets/d_score.png"/>

### Cluster A Score

<img src="assets/a_score.png"/>

## The SpinachYAML Configuration

NOTE: This file will change as Popeye matures!

```yaml
# A Popeye sample configuration file
popeye:
  # Configure node resources.
  node:
    # Limits set a cpu/mem threshold in % ie if cpu|mem > limit a lint warning is triggered.
    limits:
      # CPU checks if current CPU utilization on a node is greater than 90%.
      cpu:    90
      # Memory checks if current Memory utilization on a node is greater than 80%.
      memory: 80
    # Exclude lists node names to exclude from the scan.
    exclude:
    - master

  # Configure namespace resources
  namespace:
    # Exclude list out namespaces to be excluded from the scan.
    exclude:
      - kube-system
      - kube-public

  # Configure pod resources
  pod:
    # Restarts check the restarts count and triggers a lint warning if above threshold.
    restarts:
      3
    # Check container resource utilization in percent.
    # Issues a lint warning if about these threshold.
    limits:
      cpu:    80
      memory: 75

  # Service ...
  service:
    # Excludes these services from the scan.
    exclude:
      - default/kubernetes
      - blee-ns/fred
```

## Supported Resources

This initial drop only supports a handful of resources. More will be added soon...

* Node
* Namespace
* Pod
* Service

## Known Issues

This initial drop is brittle. Popeye will most likely blow up...

* You're running older versions of Kubernetes. Popeye works best Kubernetes 1.13+.
* You don't have enough RBAC fu to manage your cluster (see RBAC section below)
* Your cluster does not run a metric server.

## RBAC POW!

In order for Popeye to do his work, the signed in user must have enough oomph to
get/list the resources mentioned above as well as metrics-server get/list access.

## Disclaimer

This is work in progress! If there is enough interest in the Kubernetes
community, we will enhance per your recommendations/contributions. Also if you
dig this effort, please let us know that too!

## ATTA Girls/Boys!

Popeye sits on top of many of opensource projects and libraries. Our *sincere*
appreciations to all the OSS contributors that work nights and weekends
to make this project a reality!

## Contact Info

1. **Email**:   fernand@imhotep.io
2. **Twitter**: [@kitesurfer](https://twitter.com/kitesurfer?lang=en)

---

<img src="assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC.
All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
