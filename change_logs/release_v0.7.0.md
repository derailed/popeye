<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye_logo.png" align="right" width="200" height="auto"/>

# Release v0.7.0

## Notes

As you may have noticed this project offers a GitHub Sponsor button (over here 👆). If you feel `Popeye` sanitizers are helping you diagnose potential cluster issues and it's saving you some cycles, you may consider sponsoring this project. Thank you for your gesture of kindness and for supporting Popeye!!

Can't afford it but still dig this tool? Please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### HTML Reports!

Big thanks in this release and ATTA Boy! goes to (Barak Stout)[https://github.com/BarakStout] for making HTML sanitizers reports available for all of us.

In order to start spewing out HTML formatted Popeye's report use the following command:

```shell
# Generates a Popeye sanitizer HTML report
popeye -o html --save
```

And you'll get something like this...

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/html_report.png" align="right" width="500" height="auto"/>


### Cluster Name Override

In most cases, when running Popeye in cluster, the cluster name won't be available as a kubeconfig is generally not available. Thanks to the great contributions from [Karan Magdani](https://github.com/karanmagdani1), you can now add
a cli arg `k8s-popeye-cluster-name` to set your cluster name for your in cluster sanitizer reports.

---

## Resolved Bugs/PRs

* [Issue #73](https://github.com/derailed/popeye/issues/73)
* [PR 81](https://github.com/derailed/popeye/pull/81)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
