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

Install Helm: https://helm.sh/docs/intro/install/  

**The following are optional but highly recommended:**  
Install kubectl: https://kubernetes.io/docs/tasks/tools/  

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

`task apply-config`

`task bootstrap`

`task kubeconfig`

`task install-cilium`