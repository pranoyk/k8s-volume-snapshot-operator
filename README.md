# k8s-volume-snapshot-operator

The aim of this assignment is to create a kubernetes operator that takes a volume snapshot and later creates a pvc based out of it.

In order to set it up locally you have to set up a CSI driver as described in this.
https://github.com/kubernetes-csi/csi-driver-host-path/blob/master/docs/example-snapshots-1.17-and-later.md

## Usage 

To create a volume snapshot
```bash
go run main.go -kubeconfig=<path-to-kube-config> -action=createSnapshot -pvc=<existing-pvc-name> -snapshot=<snapshot-name>
```

To create a pvc out of existing volume snapshot
```bash
go run main.go -kubeconfig=<path-to-kube-config> -action=createPVCFromSnapshot -pvc=<new-pvc-name> -snapshot=<existing-snapshot-name>
```