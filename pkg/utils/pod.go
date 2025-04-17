package utils

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func AllActivePods(clientSet kubernetes.Interface) ([]v1.Pod, error) {
	allPods, err := clientSet.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "status.phase=Running",
	})
	if err != nil {
		return nil, err
	}
	return allPods.Items, nil
}

func GpuInPod(pod *v1.Pod, rName v1.ResourceName) (gpuCount int64) {
	containers := pod.Spec.Containers
	for _, container := range containers {
		val, ok := container.Resources.Limits[rName]
		if !ok {
			continue
		}
		gpuCount += val.Value()
	}
	return gpuCount
}
