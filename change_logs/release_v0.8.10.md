<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye_logo.png" align="right" width="200" height="auto"/>

# Release v0.8.10

## Notes

Thank you to all that contributed with flushing out issues and enhancements for Popeye! I'll try to mark some of these issues as fixed. But if you don't mind grab the latest rev and see if we're happier with some of the fixes! If you've filed an issue please help me verify and close. Your support, kindness and awesome suggestions to make Popeye better is as ever very much noticed and appreciated!

This project offers a GitHub Sponsor button (over here 👆). If you feel `Popeye` sanitizers are helping you diagnose potential cluster issues and it's saving you some cycles, please consider sponsoring this project!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## A Word From Our `Meager Sponsor...

Contrarily to popular belief OSS is not free! If you find Popeye useful to your organization and saves you some cycles while building/debugging your clusters and want to see this project grow and succeed, please consider becoming a sponsor member and hit that github button already!

## Change Logs

### Non zero exit

The Popeye binary will now return a non zero exit status if errors are raised during a cluster scan. This provides an affordance for CI/CD pipeline scripts to fail a deployment based on the scan reports outcomes.

---

## Resolved Bugs/PRs

- [0.8.9 changes exit status behavior in a breaking way](https://github.com/derailed/popeye/issues/122)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2020 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
