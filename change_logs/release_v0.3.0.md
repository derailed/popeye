<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.3.0

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### New Sanitizers

Added Sanitizer reports for the following resources:

+ Deployment
+ StatefulSet
+ HorizontalPodAutoscaler
+ PersistentVolume
+ PersistentVolumeClaim

Popeye will now scan for configuration and usage issues that may arise from these resources.

### WARNING! Capacitors are Charged Up!

Ever wondered how much cluster capacity you actually need? Or which resource scaling may cause your cluster to surpass it's capacity? Fear not my friends! In this release, we introduce `Capactor`. We've added metrics monitoring to the sanitizer reports. The `Capacitor` checks your resources (provided they are set!) for potential over/under allocation based on reported metrics. Additionally, the `Capacitor` checks your HorizontalPodAutoscalers and pre-computes resource allocations based on max replicas. Thus you can be warned when there is a potential for your clusters to either reach or surpass their capacity.

Mind you, this is very much still experimental, so procceed with caution!

### Report Formats

Added support for YAML and JSON output via `-o` CLI parameter.

> NOTE! Jurassic mode, though still in full effect, has been moved to `-o jurassic`

---

## Resolved Bugs

+ [Issue #22](https://github.com/derailed/popeye/issues/22)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/> © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
