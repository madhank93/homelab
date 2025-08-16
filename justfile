# justfile

# proxmox-ip := `yq '.proxmox.endpoint' ./infra/config.yml | grep -oE '[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+'`

[working-directory: 'infra']
deploy:
    pulumi up

[working-directory: 'infra']
preview:
    pulumi preview

[working-directory: 'infra']
destroy:
    pulumi destroy

[working-directory: 'infra']
refresh:
    pulumi refresh

ping_scan:
    nmap -sn 192.168.1.0/24

run-kubespray:
    docker run --rm -it --mount type=bind,source="$(pwd)/platform/k8s_cluster_config",dst=/config \
      --mount type=bind,source="${HOME}/.ssh/id_ed25519",dst=/root/.ssh/id_ed25519 \
      quay.io/kubespray/kubespray:v2.28.0 bash -c "ansible-playbook -i /config/inventory/hosts.ini -e @/config/values.yml cluster.yml"

reset-kubespray:
    docker run --rm -it --mount type=bind,source="$(pwd)/platform/k8s_cluster_config",dst=/config \
      --mount type=bind,source="${HOME}/.ssh/id_ed25519",dst=/root/.ssh/id_ed25519 \
      quay.io/kubespray/kubespray:v2.28.0 bash -c "ansible-playbook -i /config/inventory/hosts.ini -e @/config/values.yml reset.yml"

install-addons addon_tags:
    docker run --rm -it --mount type=bind,source="$(pwd)/platform/k8s_cluster_config",dst=/config \
      --mount type=bind,source="${HOME}/.ssh/id_ed25519",dst=/root/.ssh/id_ed25519 \
      quay.io/kubespray/kubespray:v2.28.0 bash -c "ansible-playbook -i /config/inventory/hosts.ini -e @/config/values.yml cluster.yml --tags {{addon_tags}}"

[working-directory: 'platform/cdk8s']
synth:
    @go mod tidy
    @cdk8s synth --output app

# stop-vm vm_id:
#     ssh root@{{proxmox-ip}} "rm -f /run/lock/qemu-server/lock-{{vm_id}}.conf && qm unlock {{vm_id}} && qm stop {{vm_id}}"
