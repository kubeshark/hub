package kubernetes

import (
	"github.com/kubeshark/worker/models"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getMinimizedContainerStatuses(fullPod v1.Pod) []v1.ContainerStatus {
	result := make([]v1.ContainerStatus, len(fullPod.Status.ContainerStatuses))

	for i, container := range fullPod.Status.ContainerStatuses {
		result[i] = v1.ContainerStatus{
			ContainerID: container.ContainerID,
		}
	}

	return result
}

func getMinimizedPod(fullPod v1.Pod) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fullPod.Name,
			Namespace: fullPod.Namespace,
		},
		Status: v1.PodStatus{
			PodIP:             fullPod.Status.PodIP,
			ContainerStatuses: getMinimizedContainerStatuses(fullPod),
		},
	}
}

func GetNodeHostToTappedPodsMap(tappedPods []v1.Pod) models.NodeToPodsMap {
	nodeToTappedPodMap := make(models.NodeToPodsMap)
	for _, pod := range tappedPods {
		minimizedPod := getMinimizedPod(pod)

		existingList := nodeToTappedPodMap[pod.Spec.NodeName]
		if existingList == nil {
			nodeToTappedPodMap[pod.Spec.NodeName] = []v1.Pod{minimizedPod}
		} else {
			nodeToTappedPodMap[pod.Spec.NodeName] = append(nodeToTappedPodMap[pod.Spec.NodeName], minimizedPod)
		}
	}
	return nodeToTappedPodMap
}

func GetPodInfosForPods(pods []v1.Pod) []*models.PodInfo {
	podInfos := make([]*models.PodInfo, 0)
	for _, pod := range pods {
		podInfos = append(podInfos, &models.PodInfo{Name: pod.Name, Namespace: pod.Namespace, NodeName: pod.Spec.NodeName})
	}
	return podInfos
}
