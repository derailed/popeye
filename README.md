<img src="assets/popeye_boat.png" align="right" width="250" heigh="auto">

# Popeye -A Kubernetes Linter


Popeye is a utility that lints a K8s cluster and reports potential issues with
various Kubernetes resources. It cruises thru deployed resources for potential misconfigurations and scans a cluster to ensure best practices are in place thus
preventing potential future headaches. It aim at reducing the cognitive *over*load
that one faces when managing and operating a Kubernetes cluster in the wild.

<br/>
<br/>

---

## Installation

---

## The Command Line

---

## Supported Resources

This initial drop only supports a handful of resources at this time. More to come soon...

- Node
- Namespace
- Pod
- Service

---

## Known Issues

This initial drop is brittle. Popeye will most likely blow up...

- You're running older versions of Kubernetes. Popeye works best Kubernetes 1.13+.
- You don't have enough RBAC fu to manage your cluster (see RBAC section below)
- Your cluster does not run a metric server.

---

## RBAC POW

In order for Popeye to do his work, the signed in user must have enough oomph to
list and get the resources mentioned above as well as metrics-server get/list access.

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
