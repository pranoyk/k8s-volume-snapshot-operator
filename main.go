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

const (
	CreateSnapshot        = "createSnapshot"
	CreatePVCFromSnapshot = "createPVCFromSnapshot"
	SnapshotClassName     = "csi-hostpath-snapclass"
	StorageClassName      = "csi-hostpath-sc"
)

func main() {
	action := flag.String("action", "", fmt.Sprintf("action to perform, %s or %s", CreateSnapshot, CreatePVCFromSnapshot))
	kubeconfig := flag.String("kubeconfig", "", "location of your kubeconfig file")
	pvcName := flag.String("pvc", "", "name of the PVC")
	snapshotName := flag.String("snapshot", "", "name of the snapshot")
	flag.Parse()

	if pvcName == nil || *pvcName == "" {
		log.Fatalf("PVC name cannot be empty")
	}

	if snapshotName == nil || *snapshotName == "" {
		log.Fatalf("Snapshot name cannot be empty")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig %s, Error: %v", *kubeconfig, err)
	}

	snapshotClient, err := snapshotclientset.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating snapshot client: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}

	switch *action {
	case CreateSnapshot:
		fmt.Printf("Creating snapshot, pvcName: %s snapshotName: %s\n", *pvcName, *snapshotName)
		snapshotName, err := createSnapshot(snapshotClient, pvcName, *snapshotName, SnapshotClassName)
		if err != nil {
			log.Fatalf("Error creating snapshot: %v", err)
		}
		fmt.Printf("Created snapshot: %s\n", snapshotName)
	case CreatePVCFromSnapshot:
		fmt.Printf("Creating PVC from snapshot: %s, pvcName: %s\n", *snapshotName, *pvcName)
		err := createPVCFromSnapshot(clientset, *snapshotName, *pvcName, StorageClassName)
		if err != nil {
			log.Fatalf("Error creating PVC from snapshot: %v", err)
		}
	default:
		log.Fatalf("Invalid action %s, action must be one of %v", *action, []string{CreateSnapshot, CreatePVCFromSnapshot})
	}
}

func createSnapshot(snapshotClient *snapshotclientset.Clientset, pvcName *string, snapshotName string, snapshotClassName string) (string, error) {
	snapshot := &snapshotsv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name: snapshotName,
		},
		Spec: snapshotsv1.VolumeSnapshotSpec{
			VolumeSnapshotClassName: &snapshotClassName,
			Source: snapshotsv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: pvcName,
			},
		},
	}
	volumeSnapshot, err := snapshotClient.SnapshotV1().VolumeSnapshots("default").Create(context.TODO(), snapshot, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("Error creating snapshot: %v", err)
	}
	return volumeSnapshot.Name, nil
}

func createPVCFromSnapshot(clientset *kubernetes.Clientset, snapshotName, pvcName, storageClassName string) error {
	apiGroup := "snapshot.storage.k8s.io"
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: pvcName,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &storageClassName,
			DataSource: &corev1.TypedLocalObjectReference{
				Name:     snapshotName,
				Kind:     "VolumeSnapshot",
				APIGroup: &apiGroup,
			},
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
