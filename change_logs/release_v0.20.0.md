<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye_logo.png" align="right" width="200" height="auto"/>

# Release v0.20.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for Popeye! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make Popeye better is as ever very much noticed and appreciated!

This project offers a GitHub Sponsor button (over here 👆). As you well know this is not pimped out by big corps with deep pockets. If you feel `Popeye` is saving you cycles diagnosing potential cluster issues please consider sponsoring this project!! It does go a long way in keeping our servers lights on and beers in our fridge.

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## ♫ Sounds Behind The Release ♭

🏹💕 Happy belated Valentines 💕🏹

* [Glory Box - Portishead](https://www.youtube.com/watch?v=NVuRbwnav_Y)
* [Funny Valentine - Elvis Costello](https://www.youtube.com/watch?v=ni3DjM8wcds)
* [Cause We've Ended As Lovers - Jeff Beck](https://www.youtube.com/watch?v=VC02wGj5gPw)

---

## 🎉 Feature Release 🥳

Popeye just got a new spinach formula and pipe!

😳 This is a big one! 😳

> NOTE! 🫣 Paint is still fresh on this deal and I might have broken stuff in the process ;(
> Please help us vet this drop to help us solidify and make Popeye better for all of us.
> Thank you!!

Splendid! So what changed?

### Biffs'em If You Got'em!

As of this drop, Popeye linters family got extended. The following linters were added/extended:

* Cronjobs
* Jobs
* Gateway-Classes
* Gateways
* HTTPRoutes
* NetworkPolicies (Beefed up!)

### New Spinach Formula!

The SpinachYAML configuration changed and won't be compatible with previous versions.
The new format provides for global exclusions and linters specific ones.
Please see the docs for the gory details but in short this is what a spinach file now looks like:

```yaml
popeye:
  allocations:
    cpu:
      underPercUtilization: 200
      overPercUtilization: 50
    memory:
      underPercUtilization: 200
      overPercUtilization: 50

  # [!!NEW!!] Specify global exclusions for fqn, codes, labels, annotations
  excludes:
    global:
      # Exclude kube-system ns for all linters.
      fqns: [rx:^kube-system]
      # Exclude these workload labels for all linters.
      labels:
        app: [blee, bozo]

    # [!!NEW!!] Linters exclude section
    linters:
      # [!!NEW!!] use the R from GVR resource specification to name the linter
      statefulsets:
        # [!!NEW!!] Exclude codes via regexp ie skip 101, 1000,...
        codes: ["rx:^10"]
        instances:
          # Skip scan for a particular FQN aka namespace/res-name
          - fqns: [default/prom-alertmanager]
            codes: [106]

      pods:
        codes: ["306", "rx:^11"]
        instances:
          - fqns: [rx:^default/prom]
          - fqns: [rx:^default/graf]
          # [!!NEW!!] Skip using either labels or annotations and/or specific codes
          - labels:
              app: [blee, blah, zorg]
            codes: [300]
          - fqns: [rx:^default/pappi]
            codes: [300, 102, 306]
            containers: [c1]

  resources:
    node:
      limits:
        cpu: 90
        memory: 80
    pod:
      limits:
        cpu: 80
        memory: 75
      restarts: 3

  overrides:
    - code: 1502
      severity: 3

  registries:
    - quay2.io
    - docker1.io
```

### Popeye The Prom Queen?

Additionally, we've updated Popeye's prometheus metrics to provide more scan insights and signals. Please see the docs for details.

. `popeye_severity_total` [gauge] tracks various counts based on severity.
. `popeye_code_total` [gauge] tracks counts by Popeye's linter codes.
. `popeye_linter_tally_total` [gauge] tracks counts per linters.
. `popeye_report_errors_total` [gauge] tracks scan errors totals.
. `popeye_cluster_score` [gauge] tracks scan report scores.

---

## Resolved Issues

. [#265](https://github.com/derailed/popeye/issues/265) additional/fine grained prometheus metrics
. [#237](https://github.com/derailed/popeye/issues/237) Support multiple outputs at once
. [#235](https://github.com/derailed/popeye/issues/235) --lint level does not affect html output
. [#232](https://github.com/derailed/popeye/issues/232) Metrics get overridden when using the same Pushgateway for multiple k8s clusters
. [#231](https://github.com/derailed/popeye/issues/231) wrong warning: [POP-107] No resource limits defined
. [#230](https://github.com/derailed/popeye/issues/230) APIs: metrics.k8s.io/v1beta1: the server is currently unable to handle the request
. [#214](https://github.com/derailed/popeye/issues/214) [POP-1100] No pods match service selector - should not be detected for ExternalName service type
. [#213](https://github.com/derailed/popeye/issues/213) Ingress extensions/v1beta1 deprecated (and deleted in k8s v1.22) is not detected ONLY in kube-metriques namespace
. [#212](https://github.com/derailed/popeye/issues/212) Ingress networking.k8s.io/v1beta1 deprecated since k8s v1.19 and deleted in k8s v1.22, is not detected ONLY in specific namespace name as kube-metriques
. [#209](https://github.com/derailed/popeye/issues/209) POP-403 - PodSecurityPolicy (PSP) k8s v1.21 deprecation - k8s v1.25 deletion - not detected
. [#202](https://github.com/derailed/popeye/issues/202) False positive on NetworkPolicy using a catch all namespaceSelector
. [#163](https://github.com/derailed/popeye/issues/163) popeye 0.9.0 with K8S 1.21.0 bug on PodDisruptionBudget - Wrong default API
. [#125](https://github.com/derailed/popeye/issues/125) info/error/warning messages to the metrics sent to prometheus
. [#97](https://github.com/derailed/popeye/issues/97) Add support for explicitly sanitizing jobs to popeye
. [#59](https://github.com/derailed/popeye/issues/59) StatefulSet incorrectly determines apiVersio

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2024 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
