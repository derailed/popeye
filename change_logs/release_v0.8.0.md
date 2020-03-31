<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye_logo.png" align="right" width="200" height="auto"/>

# Release v0.8.0

## Notes

Thank you to all that contributed with flushing out issues and enhancements for Popeye! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make Popeye better is as ever very much noticed and appreciated!

This project offers a GitHub Sponsor button (over here 👆). If you feel `Popeye` sanitizers are helping you diagnose potential cluster issues and it's saving you some cycles, please consider sponsoring this project.

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

Maintenance Release!

### Kubernetes 1.18

In this drop we've updated Popeye to support the latest api changes provided by kubernetes v1.18.

### Breaking Changes!

As of this release, we've changed popeye's configuration file to leverage GVRs (group/version/resource) notation in the spinach.yml file. In order to pickup your configurations you will need to update your spinach file as follows:

```yaml
# spinach.yml
popeye:
  # Excludes define rules to exampt resources from sanitization
  excludes:
    # NOTE: Change to GVR notation vs section names.
    rbac.authorization.k8s.io/v1/clusterrolebinding:
      - name: rx:system
        codes:
          - 400
    apps/v1/daemonset:
      - name: rx:kube
```

---

## Resolved Bugs/PRs

* [Issue #87](https://github.com/derailed/popeye/issues/87)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
