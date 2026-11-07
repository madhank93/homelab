# Reading `just core <stack> preview`

A refactor must preview as no changes. A feature change must preview as exactly the resources it names. Anything else is a bug in the change until explained.

## Symbols

| Symbol | Meaning | Action |
|---|---|---|
| `+` create | New resource | Expected only for resources the change adds |
| `~` update | In-place update | Check the diffed fields are the intended ones |
| `-` delete | Resource removed | Stop unless the change removes it on purpose |
| `+-` / `-+` replace | Delete and recreate (`-+` = delete first) | Stop and find the field that forces it |
| `[diff: ~field]` | Field driving the step | Compare with the code change |

## Red flags

| Preview shows | Likely cause | Fix |
|---|---|---|
| delete + create of the same object under a new name | Logical name or parent changed | Restore the name, or add `pulumi.Aliases` |
| replace of a Proxmox VM or `talos-secrets` | Input change on a protected resource; `up` would fail on `Protect` | Revert the input, or `IgnoreChanges` it if it only seeds creation |
| update on every run with no code change | Provider reads the field back differently (null vs default) | `IgnoreChanges` on that field with a comment |
| replace of `remote.Command` `bifrost-bootstrap` | Trigger hash changed (configs or secrets under `core/cloud/bifrost/`) | Expected when Bifrost config changed; otherwise find the file that changed |
| replace of a `download.FileLegacy` image | URL or schematic changed in `config/versions.yaml` | Expected for an image bump; running VMs ignore it |
| secret value visible in the diff | Input not wrapped in `pulumi.ToSecret` | Wrap it |
| SSA field-manager conflict on a Kubernetes resource | Another manager (kubectl, Argo CD) owns the field | Ask the user before forcing ownership |

## Checklist before `up`

- [ ] No unexplained `-`, `+-` or `-+`.
- [ ] Every `~` field maps to a line in the diff.
- [ ] No protected resource in a replace.
- [ ] No plaintext secret in the output.
- [ ] Stack order respected: `talos` before `platform`; `hetzner` before `authentik`.
