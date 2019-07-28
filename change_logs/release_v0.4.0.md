<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.4.0

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

If you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

I am super excited about this drop and hope you will be too! Lot's of changes and features but also more opportunities for
breakage. So please proceed with caution and please do file issues so we can all gain from the improvements.

### Popeye's Memoires...

Until now Popeye did not really handle any kind of sanitizer run histories. We've now added a `--save` option that allows
sanitizer runs to be persisted to disk. We've also added a Popeye `diff` subcommand that diffs the last 2 runs and reports
on the deltas so you can track your cluster score deltas. The saved reports are currently defaulting to yaml.

Provided Popeye has at least 2 saved sanitizer runs, it will perform at diff using the following command:

```shell
# Perform a cluster blee sanitization and persis results to disk.
popeye -A --cluster blee --save
# Runs a sanitizer diff report on a Kubernetes cluster named blee
popeye diff --cluster blee
```

NOTE: The diff could have issues if a managed pod is reincarnated into a new pod in the set. At that time Popeye may have a tough time locating the correct diff from the previous run. We've tried our best to match, but please vet the results or report a bug if that is not the case.

NOTE: Given the code flux and the current state of affairs, we can't guarantee the quality of the diffs as the sanitizer report format are still work in progress and could change in the future. Will do our best to keep things stable, but there is a good chance the diff may yield false-positive/negative in the interim. So please take this initial pass with a grain of salt.

### Codes

We've refactored the sanitizer report to now include sanitizer codes. Each report section have a different set of codes depending on the sanitization check. For instances code `POP-106 No resource defined` will now be indicated in the report. We will document the various codes, their meanings and resolutions once we've got a chance to vet the changes and make sure we're all happy with the reports!

On this note, and an interesting side effect, you can now change the code severity level in your spinchyaml file. There has been some reports, that manifested a need to change the message severity based on your cluster policies. That said, I would warn against it, as the end goal here is to come up with a set of standard best practices across all clusters. The reason I've decided to open this up a bit was so that we can zero in as a community for clusters best practices and not my sol insights. So I will ask, that if you do feel the urge to modify a sanitizer code severity, you file an issue so that we can discuss as a community and come up with the best directives so we all win.
So let's all optimize for correctness and not cluster scores. This
is a total backdoor for going from F->A by setting all severities to OK! ;)

Here is a sample spinach.yml config to override a code severity:

```yaml
# Severity: Ok: 0, Info: 1, Warn: 2, Error: 3
popeye:
  codes:
    206:
     severity: 2 # Set severity level to Warn vs Info if No pod disruption budget is set.
```

### Going Secure...

In this drop we've also added a few security rules as sanitizer checks. This is just the begining of a long journey but you should start seeing a few security checks in your reports.

As a results Popeye will notify if the following conditions are true on your clusters:

1. Running Pods' in the default serviceaccount
1. Running containers are root
1. Mounting API server certs.

We're going to be more active in this area in the next few drops so please let us know what other checks might be useful so we can prioritize accordingly.

### Mo' Resources

In this release we've added a few new resources to the sanitization pass. Some checks are still primitive we will improve on that soon.

1. DaemonSet
2. ReplicaSet
3. Ingress
4. PodSecurityPolicy
5. NetworkPolicy

### Linux Brewed!

Sadly, we're are still having issues deploying Popeye as a snap ;( Though we're hopefull these will be resolved soon, we've decided to offer a brewed version of Popeye as an alternate for our Linux friends.

```shell
brew install derailed/popeye/popeye
```

### Deprecation

Saving the best for last. As you might be aware K8s 1.6 drop is going to actually remove some resources group/version and hence operators are going to need to not only change their application manifests but also update their dependencies in the shape of Helm charts, custom resources, operator, etc... This is going to most likely cause some serious disturbances in the force.

No worries Popeye has your back!

In this drop, we've added some very basic check for potential use of deprecated API that are going away. Since Popeye looks at a live cluster and what is actually deploy, the sanitizers will alert you of potential deprecation problems before you update your entire Kubernetes cluster to 1.6 or in having to spin up a 1.6 cluster.

To be fair, we are not entirely certain Popeye will pick up or correctly report on all issue, but it should do a fairly decent job in picking up all resources that were `kubectl` applied. Potentially Popeye may fail on operator or CRDS that potentially spin off controlled resources.

That said, we expect Popeye to report on the following deprecated API that will soon be removed in your updated clusters:

1. DaemonSet, Deployment, StatefulSet, ReplicaSet
2. Ingress
3. PodSecurityPolicy
4. NetworkPolicy

We hope you will find this feature useful and timely in helping in the migration.

I think that's a wrap for this drop. Please be mindful that a lot of code changes happened here and some breakage may occur. Please help us zeron in and file issues so we can address and make this tool more helpful to all our friends in the Kubernetes community.

---

## Resolved Bugs

* [Issue #43](https://github.com/derailed/popeye/issues/43)
* [Issue #42](https://github.com/derailed/popeye/issues/42)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
