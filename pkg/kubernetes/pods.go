package kubernetes

import (
	"github.com/kubeshark/base/pkg/models"
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

func GetNodeHostToTargettedPodsMap(targettedPods []v1.Pod) models.NodeToPodsMap {
	nodeToTargettedPodsMap := make(models.NodeToPodsMap)
	for _, pod := range targettedPods {
		minimizedPod := getMinimizedPod(pod)

		existingList := nodeToTargettedPodsMap[pod.Spec.NodeName]
		if existingList == nil {
			nodeToTargettedPodsMap[pod.Spec.NodeName] = []v1.Pod{minimizedPod}
		} else {
			nodeToTargettedPodsMap[pod.Spec.NodeName] = append(nodeToTargettedPodsMap[pod.Spec.NodeName], minimizedPod)
		}
	}
	return nodeToTargettedPodsMap
}

func GetPodInfosForPods(pods []v1.Pod) []*models.PodInfo {
	podInfos := make([]*models.PodInfo, 0)
	for _, pod := range pods {
		podInfos = append(podInfos, &models.PodInfo{Name: pod.Name, Namespace: pod.Namespace, NodeName: pod.Spec.NodeName})
	}
	return podInfos
}
