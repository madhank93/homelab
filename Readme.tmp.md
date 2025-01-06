```
VENVDIR=kubespray-venv
python3 -m venv $VENVDIR
source $VENVDIR/bin/activate
pip install -U -r kubespray/requirements.txt
```

```
pip install -U -r kubespray/contrib/inventory_builder/requirements.txt
CONFIG_FILE=inventory/hosts.yml python3 kubespray/contrib/inventory_builder/inventory.py \
  k8s-controller1,192.168.1.224 \
  k8s-controller2,192.168.1.226 \
  k8s-controller3,192.168.1.221 \
  k8s-worker1,192.168.1.253 \
  k8s-worker2,192.168.1.210 \
  k8s-worker3,192.168.1.211 \
```