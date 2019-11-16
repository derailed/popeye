<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.5.0

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

If you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

In this drop, we've cleaned up a few code duds and addressed a bit of debt.

### Prometheus Report

Thanks to an awesome contribution by [dardanel](https://github.com/eminugurkenar), Popeye can now report sanitization issues as Prometheus metrics. Thus, you will have the ability to run Popeye in cluster as a job and push sanitization metrics back to the prometheus mothership. How cool is that? As it stands these will just be reported as raw counts and thus you won't have sanitization details but you can leverage Prometheus AlertManager to trigger your clusters investigation based on these reports.

---

## Resolved Bugs

* [Issue #58](https://github.com/derailed/popeye/issues/58) Thank you @renan!!

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
