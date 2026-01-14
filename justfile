# Variables
config_path := "platform/ansible"
kubespray_version := "v2.28.0"
kubespray_image := "quay.io/kubespray/kubespray:" + kubespray_version

help:
    @just --list

##################
##### Pulumi #####
##################

# Allowed actions for Pulumi
allowed_actions := "up preview destroy refresh"

pulumi_dir := "infra/pulumi"

[working-directory: 'infra/pulumi']
pulumi stack action:
    #!/usr/bin/env bash
    set -euo pipefail

    # Validate action
    case "{{action}}" in
      up|preview|destroy|refresh) ;;
      *)
        echo "Usage: just pulumi <stack> <action>"
        echo "  <action> = up | preview | destroy | refresh"
        exit 1
        ;;
    esac

    # Validate stack from Pulumi.*.yaml
    STACK_FILE="Pulumi.{{stack}}.yaml"
    if [[ ! -f "${STACK_FILE}" ]]; then
      echo "❌ Invalid stack: '{{stack}}'"
      echo "   No file '${STACK_FILE}' found in $(pwd)"
      echo "   Available stacks:"
      ls Pulumi.*.yaml | sed 's/^Pulumi.\(.*\)\.yaml$/  - \1/'
      exit 1
    fi

    # Compute flags
    FLAGS=$([[ "{{action}}" != "preview" ]] && echo "--yes" || echo "")

    echo "🚀 [{{stack}}] pulumi {{action}} ${FLAGS}"
    pulumi stack select "{{stack}}"
    pulumi "{{action}}" ${FLAGS}


#######################
##### Networking ###### 
#######################
ping-scan:
	nmap -sn 192.168.1.0/24

######################
##### kubespray ######
######################
base_docker_cmd := "docker run --rm -it " + \
  "--mount type=bind,source=${LOCAL_WORKSPACE_FOLDER}/" + config_path + ",target=/config " + \
  "--mount type=bind,source=${HOST_HOME}/.ssh/id_ed25519,target=/root/.ssh/id_ed25519,readonly " + \
  kubespray_image

run-kubespray:
	{{base_docker_cmd}} ansible-playbook -i /config/inventory/hosts.ini -e @/config/values.yml cluster.yml

reset-kubespray:
	{{base_docker_cmd}} ansible-playbook -i /config/inventory/hosts.ini -e @/config/values.yml reset.yml

install-addons addon_tags:
	{{base_docker_cmd}} ansible-playbook -i /config/inventory/hosts.ini -e @/config/values.yml cluster.yml --tags {{addon_tags}}

##################
##### CDK8s ######
##################
[working-directory: 'platform/cdk8s']
synth:
	go mod tidy
	cdk8s synth --output ../../app

[working-directory: 'platform/cdk8s']
apply: synth
	kubectl apply -f app

########################
#### Misecellaneous ####
########################
folder_structure:
	tree -P "*.go|*.ini|*.yml" --gitignore -I "imports|assets"