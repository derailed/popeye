<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.3.0

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### New Sanitizers

Added Sanitize reports for the following resources:

+ HorizontalPodAutoscaler
+ PersistentVolume
+ PersistentVolumeClaim
+ StatefulSet
+ Deployment

Popeye will now scan for configuration and usage issues that may arise from these resources.
In addition, Popeye now includes resource utilization tracker codename: `Capacitor` to track over/under resource allocations for cpu and memory. Furthermore the **Capacitor** warns you
when you've maxed out your cluster capacity should say HPA's start firing. This assumes your
pods declarations include resource requests/limits.

### Report Formats

Added support for YAML and JSON output via `-o` CLI parameter.

NOTE: The jurassic mode tho still available has been moved to `-o jurassic`


---

## Resolved Bugs

+ [Issue #22](https://github.com/derailed/popeye/issues/22)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
