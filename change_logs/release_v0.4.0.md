<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.4.0

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

If you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

I am super excited about this drop and hope you will be too! Lot's of changes and features but also more opportunities for
breakage. So please proceed with caution and please do file issues so we can all gain from the improvements.

### Spinach Boosts Memory?

Until now Popeye did not really handle any kind of sanitizer run histories. We've added a `--save` option that allows
sanitizer runs to be persisted to disk.

```shell
# Perform a cluster blee sanitization and persists results to disk.
popeye -A  --save
```

### Junit Output

Some folks had requested a junit flavored output for integration with CI/CD tools like Jenkins. To this end, we've provided a new formatter to output sanitizer reports as Junit flavored XML.

In order to enable the report, use the following argument:

```shell
popeye -o junit
```

NOTE: This is an experimental feature and subject to change based on users feedback!

### Codes

We've refactored the sanitizer report to now include sanitizer codes. Each report section have a different set of codes depending on the sanitization checks. For instance, code `POP-106 No resource defined` will now be indicated in the report. We will document the various codes, their meanings and resolutions once we've got a chance to vet the changes and make sure we're all happy with the new reports!

On this note, and an interesting side effect, you can now change the code severity level in your `spinach` config file. There has been some reports, voicing a need to change the message severity based on your cluster policies. That said, I would warn against it, as the end goal here is to come up with a set of standard best practices across all clusters. The reason we' ve decided to open this up a bit was so that we can zero in as a community for clusters best practices. So I will ask, that if you do feel the urge to modify a sanitizer code severity, you file an issue so that we can discuss as a group and come up with the best directives so we can all endup with a winner. This is a total backdoor for improving your clusters score without changing any manifests...

Here is a sample spinach.yml config to override a code severity:

```yaml
# Severities: Ok: 0, Info: 1, Warn: 2, Error: 3
popeye:
  codes:
    206:
     severity: 2 # Set severity level to Warn vs Info if No pod DisruptionBudget is set.
```

### Security Now!

In this drop we've also added a few security rules as sanitizer checks. This is just the begining of a long journey but you should start seeing a few security checks in your reports.

As a results Popeye will notify if the following conditions are true on your clusters:

1. Running Pods using the `default` ServiceAccount
2. Running containers are root
3. Warning about mounting API server certs on pods.

We're going to be more active in this area in the next few drops so please let us know which checks might be most useful so we can prioritize accordingly.

### Mo' Resources

In this release we've added a few new resources to the sanitization pass. Some checks are still primitive we will improve on that soon.

1. DaemonSet
2. ReplicaSet
3. Ingress
4. PodSecurityPolicy
5. NetworkPolicy

### Linux Brewed!

Sadly, we're are still having issues deploying Popeye as a snap ;( Though we're hopeful these will be resolved soon, we've decided to offer a brewed version of Popeye as an alternate for our [Linux](https://docs.brew.sh/Homebrew-on-Linux) friends.

```shell
brew install derailed/popeye/popeye
```

### 1.6 Deprecations

Saving the best for last! As you might be aware K8s 1.6 release is going to remove some resource api group version in the schema. Cluster admins/operators are going to need to not only change their application manifests but also update their dependencies. This is going to most likely cause some disturbance in the force. No worries Popeye has your back!

In this drop, we've added some very basic checks for potential use of the deprecated APIs. Since Popeye looks at a live cluster and what is actually deployed and running, the sanitizers will alert you of potential deprecation problems before you update your entire Kubernetes cluster to 1.6.

Popeye sanitizers will warn you on deprecated resource api groups on the following:

1. extensions/v1beta1 or apps/v1beta1 or apps/v1beta2 for DaemonSet, Deployment, StatefulSet, ReplicaSet
2. extensions/v1beta1.Ingress
3. extensions/v1beta1.PodSecurityPolicy
4. extensions/v1beta1.NetworkPolicy

NOTE: Is it possible that Popeye might not cover 100% of the cases as Helm charts or operators implementation might bypass the basic checks Popeye is relying on to determine a resource api group version.

We hope you will find these features useful and timely in helping in the migration.

I think that's a wrap for this drop. Please be mindful that a lot of code changes happened here and some breakage might occur. Please help us zero in and file issues should you experience incorrect reports. Thank you!!

---

## Resolved Bugs

* [Issue #43](https://github.com/derailed/popeye/issues/43)
* [Issue #42](https://github.com/derailed/popeye/issues/42)
* [Issue #35](https://github.com/derailed/popeye/issues/35)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
