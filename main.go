package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	snapshotsv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	snapshotclientset "github.com/kubernetes-csi/external-snapshotter/client/v4/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	kubeconfig := flag.String("kubeconfig", "/path/to/kubeconfig", "location of your kubeconfig file")
	pvcName := flag.String("pvc", "", "name of the PVC")
	snapshotName := flag.String("snapshot", "", "name of the snapshot")
	flag.Parse()

	fmt.Printf("Creating snapshot %s for PVC %s\n", *snapshotName, *pvcName)
	config, _ := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	snapshotClient, _ := snapshotclientset.NewForConfig(config)

	// Create a snapshot
	snapshot := &snapshotsv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      *snapshotName,
			Namespace: "default",
		},
		Spec: snapshotsv1.VolumeSnapshotSpec{
			Source: snapshotsv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: pvcName,
			},
		},
	}

	volumeSnapshot, err := snapshotClient.SnapshotV1().VolumeSnapshots("default").Create(context.TODO(), snapshot, metav1.CreateOptions{})
	if err != nil {
		log.Fatalf("Error creating snapshot: %v", err)
	}
	fmt.Printf("Created snapshot: %s\n", volumeSnapshot.Name)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}
	createPVCFromSnapshot(clientset, "pvc-from-snapshot", "manual-snapshot")
}

func createPVCFromSnapshot(clientset *kubernetes.Clientset, pvcName, storageClassName string) error {
	// Create a PVC from the snapshot
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: "default",
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
		},
	}

	pvc, err := clientset.CoreV1().PersistentVolumeClaims("default").Create(context.TODO(), pvc, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Error creating PVC: %v", err)
	}
	fmt.Printf("Created PVC: %s\n", pvc.Name)
	return nil
}
