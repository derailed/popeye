<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye_logo.png" align="right" width="200" height="auto"/>

# Release v0.8.9

## Notes

Thank you to all that contributed with flushing out issues and enhancements for Popeye! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make Popeye better is as ever very much noticed and appreciated!

This project offers a GitHub Sponsor button (over here 👆). If you feel `Popeye` sanitizers are helping you diagnose potential cluster issues and it's saving you some cycles, please consider sponsoring this project!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## A Word From Our `Meager Sponsor...

Contrarily to popular belief OSS is not free! If you find Popeye useful to your organization and save you some cycles while building/debugging your clusters and want to see this project grow and succeed, please consider becoming a sponsor member and hit that github button already!

## Change Logs

### Non zero exit

The Popeye binary will now return a non zero exit status if warnings or errors are raised during a cluster scan. This provides an affordance for CI/CD pipeline scripts to fail a deployment based on the scan reports outcomes.

### Excluding containers from report

In this drop, we've added a new spinach option to exclude certain containers for the scans. Here is a an example of a spinach file that will omit given containers from the scans.

```yaml
popeye:
  # Excludes define rules to exempt resources from scans
  excludes:
    v1/pods:
      - name: rx:icx/.*
        containers:
          # Excludes istio init and sidecar container from scan!
          - istio-proxy
          - istio-init
```

---

## Resolved Bugs/PRs

- [Extend excludes to allow excluding specific containers of a pod](https://github.com/derailed/popeye/issues/120)
- [Deamonset pods are reporting missing PodDisruptionBudget](https://github.com/derailed/popeye/issues/117)
- [Erroneous NetworkPolicy scan report on non matching pods](https://github.com/derailed/popeye/issues/116)
- [Is it possible to use the popeye command to fail a CI/CD build if there are warning or errors exists](https://github.com/derailed/popeye/issues/98)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
