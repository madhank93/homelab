# Variables
config_path := "platform/k8s_cluster_config"
kubespray_version := "v2.28.0"
kubespray_image := "quay.io/kubespray/kubespray:" + kubespray_version

##################
##### Pulumi #####
##################
[working-directory: 'infra']
deploy:
	pulumi up --yes

[working-directory: 'infra']
preview:
	pulumi preview

[working-directory: 'infra']
destroy:
	pulumi destroy --yes

[working-directory: 'infra']
refresh:
	pulumi refresh

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