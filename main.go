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
	action := flag.String("action", "", "action to perform, createSnapshot or createPVCFromSnapshot or listSnapshot")
	kubeconfig := flag.String("kubeconfig", "", "location of your kubeconfig file")
	pvcName := flag.String("pvc", "", "name of the PVC")
	snapshotName := flag.String("snapshot", "", "name of the snapshot")
	flag.Parse()

	if !isValidAction(*action) {
		log.Fatalf("Invalid action %s, action must be one of %v", *action, []string{"createSnapshot", "createPVCFromSnapshot", "listSnapshot"})
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig %s, Error: %v", *kubeconfig, err)
	}

	if *action == "createSnapshot" {
		fmt.Printf("Creating snapshot, pvcName: %s snapshotName: %s\n", *pvcName, *snapshotName)
		snapshotClient, err := snapshotclientset.NewForConfig(config)
		if err != nil {
			log.Fatalf("Error creating snapshot client: %v", err)
		}
		snapshotName, err := createSnapshot(snapshotClient, pvcName, *snapshotName)
		if err != nil {
			log.Fatalf("Error creating snapshot: %v", err)
		}
		fmt.Printf("Created snapshot: %s\n", snapshotName)
		return
	}

	if *action == "listSnapshot" {
		fmt.Println("Listing snapshots")
		snapshotClient, err := snapshotclientset.NewForConfig(config)
		if err != nil {
			log.Fatalf("Error creating snapshot client: %v", err)
		}
		snapshots, err := listSnapshot(snapshotClient)
		if err != nil {
			log.Fatalf("Error listing snapshots: %v", err)
		}
		for _, snapshot := range snapshots.Items {
			fmt.Printf("Snapshot name: %s Source PVC: %s\n", snapshot.Name, *snapshot.Spec.Source.PersistentVolumeClaimName)
		}
		return
	}

	fmt.Printf("Creating PVC from snapshot: %s, pvcName: %s\n", *snapshotName, *pvcName)
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Error creating clientset: %v", err)
	}
	createPVCFromSnapshot(clientset, *pvcName, "manual-snapshot")
}

func isValidAction(action string) bool {
	return action == "createSnapshot" || action == "createPVCFromSnapshot" || action == "listSnapshot"
}

func createSnapshot(snapshotClient *snapshotclientset.Clientset, pvcName *string, snapshotName string) (string, error) {
	snapshot := &snapshotsv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      snapshotName,
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
		return "", fmt.Errorf("Error creating snapshot: %v", err)
	}
	return volumeSnapshot.Name, nil
}

func listSnapshot(snapshotClient *snapshotclientset.Clientset) (*snapshotsv1.VolumeSnapshotList, error) {
	snapshots, err := snapshotClient.SnapshotV1().VolumeSnapshots("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("Error listing snapshots: %v", err)
	}
	return snapshots, nil
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
