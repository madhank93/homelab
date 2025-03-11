```sh
VENVDIR=kubespray-venv
python3 -m venv $VENVDIR
source $VENVDIR/bin/activate
pip install -U -r requirements.txt
```

Not available in the master branch
```sh
pip install -U -r kubespray/contrib/inventory_builder/requirements.txt
CONFIG_FILE=inventory/hosts.yml python3 kubespray/contrib/inventory_builder/inventory.py \
  k8s-controller1,192.168.1.185 \
  k8s-controller2,192.168.1.150 \
  k8s-worker1,192.168.1.182 \
  k8s-worker2,192.168.1.181 \
  k8s-worker3,192.168.1.184 \
  k8s-worker4,192.168.1.232 \

```

```sh
ansible-playbook -i ../inventory/hosts.yml -e @../values.yml --user=ubuntu --become --become-user=root cluster.yml
```

To check the available IP range in your home network using the

```sh
nmap -sn 192.168.1.0/24
```