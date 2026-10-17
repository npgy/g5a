# g5a

*granola*

Opinionated cluster setup for a single bare-metal node
- Talos
- Gateway API
- Metrics Server
- Cilium (w/ kubeproxy replacement)
- FluxCD, via Flux Operator
- Cert Manager
- SOPS

## Setup this repo:

Install go at https://go.dev/dl/  
Install task: `go install github.com/go-task/task/v3/cmd/task@latest`

Copy `data.env.example` to `data.env` and set your variables

## Image Talos on your node(s)
boot the machine in rescue mode.

generate image at factory.talos.dev

curl -LO https://url/to/metal-amd64.raw.zst

`sudo apt install zstd`

`unzstd metal-amd64.raw.zst`

run `lsblk` to view the block devices, find the appropriate one to write this image to

run `wipefs --all /dev/sdX` for necessary disks to uninitialize them

`dd if=metal-amd64.raw.zst of=/dev/sda bs=4M status=progress`
`sync`
set server to boot from hd (disable rescue mode)
`reboot`

## Build the Cluster

Run `task build` to build your cluster given the values in `data.env` using the `templates/`

This will output a folder called `infra`, which can be safely initialized as a git repo and commited to your own cluster repo.  
You may still want this repo private, as the CONTROL_PLANE_IP is in the rendered assets, however all the certs and secrets will be gitignored.  

`talosctl apply-config --insecure --nodes $CONTROL_PLANE_IP --file infra/cluster/.clusterconfig/controlplane.yaml`

`talosctl bootstrap --nodes $CONTROL_PLANE_IP`

`talosctl kubeconfig --nodes $CONTROL_PLANE_IP`



```
helm install \
    cilium \
    cilium/cilium \
    --version 1.20.2 \
    --namespace kube-system \
    --set ipam.mode=kubernetes \
    --set kubeProxyReplacement=true \
    --set securityContext.capabilities.ciliumAgent="{CHOWN,KILL,NET_ADMIN,NET_RAW,IPC_LOCK,SYS_ADMIN,SYS_RESOURCE,DAC_OVERRIDE,FOWNER,SETGID,SETUID}" \
    --set securityContext.capabilities.cleanCiliumState="{NET_ADMIN,SYS_ADMIN,SYS_RESOURCE}" \
    --set cgroup.autoMount.enabled=false \
    --set cgroup.hostRoot=/sys/fs/cgroup \
    --set k8sServiceHost=localhost \
    --set k8sServicePort=7445 \
    --set gatewayAPI.enabled=true \
    --set gatewayAPI.enableAlpn=true \
    --set gatewayAPI.enableAppProtocol=true \
    --set operator.replicas=1 \
    --set ingressController.loadbalancerMode=shared
```

Set your CONTROL_PLANE_IP in data.env or secrets.env

`task dev`

In the generated folder will be a CiliumLoadBalancerIPPool that you should now apply to your cluster
