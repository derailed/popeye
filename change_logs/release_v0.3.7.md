<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.3.7

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

If you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Spinach Config Reloaded!

BREAKING CHANGE!

As of this release the spinach.yml format has changed slightly. There is now a new `exludes` section that allows one to exclude any Kubernetes resources from the sanitizer run. A resource is identified by a resource kind and a fully qualified resource name ie `namespace/resource_name`. For example a pod named fred-1234 in namespace blee FQN will be `blee/fred-1234`. This provides for differentiating `fred/p1` and `blee/p1`. For cluster wide resources, `FQN=name`. Exclude rules can have either a straight string match or a regular expression. In the later case the regular expression must be indicated using the `rx:` prefix.

NOTE! Please thread carefully here with your regex as more resources than expected may get excluded from the report via a *loose* regex rule. When your cluster resources change, this could lead to rendering sanitization sub-optimal. Once in a while it might be a good idea to run Popeye `Config less` to make sure you're trapping any new issues with your clusters...

Here is an example spinach file as it stands in this release:

```yaml
popeye:
  allocations:
    cpu:
      over: 200
      under: 50
    memory:
      over: 200
      under: 50

  # New excludes section now provides for excluding any resources scanned by Poppeye.
  excludes:
    # Exclude any configmaps within namespace fred that ends with a version#
    configmap:
      - rx:fred*\.v\d+
    # Exclude kube-system + any namespace the start with either kube or istio
    namespace:
      - kube-public
      - rx:kube
      - rx:istio
    # Exclude node named n1 from the scan.
    node:
      - n1
    # Exclude any pods that start with nginx or contains -telemetry
    pod:
      - rx:nginx
      - rx:.*-telemetry
    # Exclude any service containing -dash in their name.
    service:
      - rx:*-dash

  # Node...
  node:
    limits:
      cpu:    90
      memory: 80

  # Pod...
  pod:
    limits:
      cpu:    80
      memory: 75
    restarts: 3
```

> NOTE: Malformed regex issues will be surfaced in the logs! Please use `popeye version` for logs location.


---

## Resolved Bugs

+ [Issue #30](https://github.com/derailed/popeye/issues/30)
+ [Issue #32](https://github.com/derailed/popeye/issues/32)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
