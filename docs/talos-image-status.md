# Talos Image Build Status Check

Run this command periodically to check if the installer image is ready:

```bash
curl -sI "https://factory.talos.dev/installer/613e1592b2da41ae5e265e8789429f22e121aab91cb4deb6bc3c0b6262961245:v1.9.3" | grep "HTTP/"
```

**When ready**: Should return `HTTP/2 200` instead of `HTTP/2 404`

**Then run**:
```bash
cd /workspace/infra/pulumi
bash ../../scripts/talos-upgrade.sh
```

## Schematic Details

**ID**: `613e1592b2da41ae5e265e8789429f22e121aab91cb4deb6bc3c0b6262961245`

**Extensions**:
- `siderolabs/iscsi-tools` (for Longhorn)
- `siderolabs/util-linux-tools` (for Longhorn)

**Note**: GPU extensions removed to speed up build time. Can add later if needed.

## Estimated Wait Time

- **First 5-10 minutes**: Image building
- **10-20 minutes**: Should be available
- **20+ minutes**: Check factory.talos.dev status

## Alternative: Check Build Progress

```bash
# Try accessing any image format to trigger build
curl -sI "https://factory.talos.dev/image/613e1592b2da41ae5e265e8789429f22e121aab91cb4deb6bc3c0b6262961245/v1.9.3/metal-amd64.iso"
```
