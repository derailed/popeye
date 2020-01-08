<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/popeye.png" align="right" width="200" height="auto"/>

# Release v0.6.0

## Notes

Thank you so much for your support and suggestions to make Popeye better!!

If you dig this tool, please make some noise on social! [@kitesurfer](https://twitter.com/kitesurfer)

---

## Change Logs

### Popeye's got your RBAC!

New this release, we've added preliminary sanitizers for the following RBAC resources: clusterrole, clusterrolebinding, role and rolebinding. The sanitizers will now check if these resource are indeed in use on your clusters.

## Excludes are OUT??

We've revamped the way excludes worked. Big thanks and credits goes to [Dirk Jablonski](https://github.com/djablonski-moia) for the push! So you can now excludes some sanitizers based not only on the resource name and type but also based on the sanitization codes. ie exclude all pod freds as long as they have missing probes (Code=102) but flag any other issues. This I think will make Popeye a bit more flexible.

NOTE: You will need to revamp your spinachYAML files as the format changed!!

Here is an example:

```yaml
popeye:
  # Excludes define rules to exempt resources from sanitization
  excludes:
    # NOTE!! excludes now use the full singular resource kind ie pod and not po or pods.
    pod:
      # Excludes all pods named fred unless the sanitizer reports any different codes from 102 or 106
      - name: rx:fred
        codes:
        - 102
        - 106
```

Please keep in mind the paint is still fresh here and I could have totally hosed some stuff in the process. If so reach out for your issues/prs button.

Thank you all for your great suggestions, fixes, patience and kindness!!

---

## Resolved Bugs

* [Issue #46](https://github.com/derailed/popeye/issues/46)
* [Issue #51](https://github.com/derailed/popeye/issues/51)
* [Issue #60](https://github.com/derailed/popeye/issues/60)
* [Issue #61](https://github.com/derailed/popeye/issues/61)
* [Issue #62](https://github.com/derailed/popeye/issues/62)

---

<img src="https://raw.githubusercontent.com/derailed/popeye/master/assets/imhotep_logo.png" width="32" height="auto"/>&nbsp; © 2019 Imhotep Software LLC. All materials licensed under [Apache v2.0](http://www.apache.org/licenses/LICENSE-2.0)
