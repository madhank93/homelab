+++
title = "Packer Images"
weight = 40
+++

# Packer Image Builder

Custom machine images are built using [Packer](https://www.packer.io/) with the QEMU builder. The configuration is located in `infra/packer/ubuntu.pkr.hcl`.

## Image Specifications

- **Source Image**: Ubuntu 24.04 Server Cloud Image (amd64)
- **Output Format**: `qcow2`
- **Output Directory**: `artifacts/`
- **Size**: 8G

## Build Process

The build process involves:
1.  **Download**: Fetches the base Ubuntu cloud image.
2.  **Conversion**: Converts and resizes the disk image using QEMU.
3.  **Provisioning**:
    - Waits for `cloud-init` to finish.
    - Cleans up `apt` caches.
    - Trims the filesystem (`fstrim`).

### Configuration

```hcl
source "qemu" "ubuntu" {
  iso_url          = var.image_url
  format           = "qcow2"
  disk_compression = true
  accelerator      = "kvm"
  headless         = true
  # ...
}
```

The resulting image is used as the template for Proxmox VMs.
