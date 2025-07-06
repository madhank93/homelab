Refer blog - 


## Troubleshooting Guide

### 1. Ensure you're using the correct Proxmox node

<details>
<summary>Show steps</summary>

- Check your `.env` file and make sure `PROXMOX_ENDPOINT` points to the correct node.
- In your code, confirm `NodeName` in `ClusterConfig` matches the node's name as shown in the Proxmox GUI.
- Example:

```go
clusterConfig := ClusterConfig{
    NodeName: "proxmox", // Must match the node name in Proxmox
    ...
}
```

text
</details>

---

### 2. Enable "Snippets" in storage configuration

<details>
<summary>Show steps</summary>

![enable-snippet](/assets/img/enable_snippet.png)

- In the Proxmox GUI, go to **Datacenter > Storage**.
- Edit your storage (e.g., `local`).
- In the **Content** field, check **Snippets**.
- Save changes.
</details>

---

### 3. Logging into a new VM

![ip-addr](/assets/img/ip_addr.png)

<details> 
<summary>Show steps</summary>
Find the VM’s IP (from Pulumi output or Proxmox GUI).

SSH into the VM:

```text
ssh ubuntu@<IP-address>
Use password: 12345
```

</details>
