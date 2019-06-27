<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.3.13

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

If you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Add over-allocs flag

Popeye is designed to report sanitization on a live cluster. As such when a cluster is mainly idle, the over allocation report may yield false positives. To this end, we've added a `--over-allocs` option to the CLI to opt-in over allocations reports. By default this option will be off, hence no over cpu/memory allocations will be reported. This now gives you an option to report allocation based on cluster load.

---

## Resolved Bugs

* [Issue #39](https://github.com/derailed/popeye/issues/39)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
