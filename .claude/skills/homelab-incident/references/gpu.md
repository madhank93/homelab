# GPU node incidents (k8s-worker4)

Pod requirements and config: [GPU](../../../../docs/src/content/docs/hardware/gpu.md).

## Node or GPU gone

The card (`0000:09:00.0`, `pcie_ids` in `config/config.yml`) sits in an external USB4 enclosure behind a Thunderbolt card on the Proxmox host. After a power loss the enclosure can stay off; Proxmox then fails to start the VM with `no PCI device found for '0000:09:00.0'` and worker4 never boots.

1. `kubectl get node k8s-worker4`: absent or `NotReady`.
2. Ask the user to run on the Proxmox host (`192.168.1.235`): `boltctl list` (expect `connected`/`authorized`) and `lspci -nn | grep 10de`.
3. `disconnected` or no NVIDIA device: enclosure power or cable. Thunderbolt authorization is not the cause.
4. Once the device is back, the user starts the VM; passthrough changes need a cold boot.
5. After boot: `kubectl get pods -n nvidia-gpu-operator`, then `kubectl describe node k8s-worker4 | grep nvidia.com/gpu`.

## GPU pods failing

| Symptom | Cause | Fix |
|---|---|---|
| `Pending`, no clear reason | Missing `dedicated=ai:NoSchedule` toleration, or all five slices allocated | Code fix in the workload (`homelab-cdk8s`) |
| Running, CUDA sees no device | Missing `runtimeClassName: nvidia` | Code fix, then recreate the pod |
| Ollama cannot load a model | ComfyUI holds VRAM | `just comfyui off` |
| ComfyUI 503 | Scaled to zero by design | `just comfyui on` |
| Longhorn volume on worker4 will not attach | Engine image missing the `dedicated=ai` toleration | [storage.md](storage.md) |

Never `kubectl scale` ComfyUI or Ollama: `selfHeal` reverts it.
