<img src="assets/popeye_boat.png" align="right" width="250" heigh="auto">

# Popeye - A Kubernetes Cluster Sanitizer

Popeye is a utility that cruises a K8s cluster resources and reports potential
issues with your deployment manifests and configurations. By scanning your
clusters, it detects misconfigurations and ensure best practices are in place thus
preventing potential future headaches. It aim at reducing the cognitive *over*load
that one faces when managing and operating a Kubernetes cluster in the wild. Popeye
is a readonly tool it does not change or update any of our Kubernetes resources or
configurations in any ways.

<br/>
<br/>

---

## Installation

Popeye is available on Linux, OSX and Windows platforms.

* Binaries for Linux, Windows and Mac are available as tarballs in the [release](https://github.com/derailed/popeye/releases) page or via the SnapCraft link above.

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
        github.com/derailed/popeye => MY_K9S_CLONED_GIT_REPO
      )
      ```

   4. Build and run the executable

        ```shell
        go run main.go
        ```

---

## The Command Line

You can use popeye standalone or using a spinach yaml config to tune the linter.
Details of the spinach yaml are below.

```shell
# Popey a cluster using your current kubeconfig environment.
popeye
# Popeye using a spinach config file
popeye -f spinach.yml
# Popeye a cluster using a kubeconfig context
popeye --cluster fred
# Stuck?
popeye help
---

## Spinach YAML

NOTE: This file will change as Popeye matures

```yaml
# A Popeye sample configuration file
popeye:
  # Configure node resources.
  node:
    # Limits set a cpu/mem threshold in % ie if cpu|mem > limit a lint warning is triggered.
    limits:
      # CPU checks if current CPU utilization on a node is greater than 80%.
      cpu:    80
      # Memory checks if current Memory utilization on a node is greater than 70%.
      memory: 70
    # Exclude lists node names to exclude from the scan.
    exclude:
    - n1

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
    # Labels NYI!! This would enforce the presence of certain labels on pods.
    labels:
    - app
    - env
```

## Supported Resources

This initial drop only supports a handful of resources. More will be added soon...

* Node
* Namespace
* Pod
* Service

---

## Known Issues

This initial drop is brittle. Popeye will most likely blow up...

* You're running older versions of Kubernetes. Popeye works best Kubernetes 1.13+.
* You don't have enough RBAC fu to manage your cluster (see RBAC section below)
* Your cluster does not run a metric server.

---

## RBAC POW!

In order for Popeye to do his work, the signed in user must have enough oomph to
get/list the resources mentioned above as well as metrics-server get/list access.

---

## Disclaimer

This is work in progress! If there is enough interest in the Kubernetes
community, we will enhance per your recommendations/contributions. Also if you
dig this effort, please let us know that too!

---

## ATTA Girls/Boys!

Popeye sits on top of many of opensource projects and libraries. Our *sincere*
appreciations to all the OSS contributors that work nights and weekends
to make this project a reality!


---

## Contact Info

1. **Email**:   fernand@imhotep.io
2. **Twitter**: [@kitesurfer](https://twitter.com/kitesurfer?lang=en)


<br/>
<br/>

---

<img src="assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC.
All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
