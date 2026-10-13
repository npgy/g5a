# g5a

*granola*

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

## Bootstrap talos

`talosctl gen secrets --output-file secrets/talos-secrets.yaml`

`talosctl gen config --with-secrets secrets/talos-secrets.yaml granola https://$CONTROL_PLANE_IP:6443 --config-patch-control-plane base/controlplane.yaml`

`talosctl gen config granola https://$CONTROL_PLANE_IP:6443 --install-disk /dev/sda`

move `talosconfig` into `~/.talos/config`

`talosctl apply-config --insecure --nodes $CONTROL_PLANE_IP --file talos-cluster/controlplane.yaml`

`talosctl config endpoints $CONTROL_PLANE_IP`

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
