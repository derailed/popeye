<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.3.6

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

Also if you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Spinach Exclude

The exclude section of the yaml now supports regular expresions. In order to designate a regular expression matcher your exclude must start with `rx:`. Here are some examples:

```yaml
exclude:
  # Exclude pod named blee.
  - blee
  # Exclude all pod name that start with nginx.
  - rx:nginx
  # Exclude all pod that contain -duh ie blee-duh and fred-duh.
  - rx:.*-duh
```

> NOTE: Malformed regex issues will be surfaced in the logs! Please use `popeye version` for logs location.

### Performance part Duh

In my speed up excitements, I've spaced checking for clusters that don't currently support metrics. This yield to an npe ;(. This should now be resolved. Sorry about this waffle thin disruption in the force!

---

## Resolved Bugs

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
