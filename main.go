package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config := flag.String("kubeconfig", "/home/pranoy/.kube/config", "Path to a kubeconfig.")
	flag.Parse()
	client, err := clientcmd.BuildConfigFromFlags("", *config)
	if err != nil {
		log.Fatalf("error while building config", err.Error())
	}
	clientset, err := kubernetes.NewForConfig(client)

	pods, err := clientset.CoreV1().Pods("default").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Println("error while listing pods", err.Error())
		return
	}
	for _, pod := range pods.Items {
		fmt.Println(pod.Name)
	}
}