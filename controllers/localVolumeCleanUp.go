package controllers

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// fetchRemovePV finds and deletes all local volume PVs
func (r *CleanUpWatcher) fetchRemovePV(ctx context.Context) bool {
	// Find PVs
	persistenceVolume := []corev1.PersistentVolume{}
	pvList := &corev1.PersistentVolumeList{}
	err := r.List(ctx, pvList)
	if err != nil {
		fmt.Print("Error in Getting PV List")
		return false
	}
	for _, pv := range pvList.Items {
		if strings.HasPrefix(pv.Name, "local-pv-") {
			persistenceVolume = append(persistenceVolume, pv)
			fmt.Println("PV status- ", pv.Status.Phase)
		}
	}
	// PV Deletion
	for _, pv := range persistenceVolume {
		err = r.Delete(ctx, &pv)
		if err != nil && !errors.IsNotFound(err) {
			fmt.Print("Error in Deleting PV ", pv.Name)
			return false
		}
	}
	fmt.Println("PV Deleted.....")
	return true
}

// deleteMountedPath deletes mounted path from each node
func (r *CleanUpWatcher) deleteMountedPath(ctx context.Context) bool {
	// Remove Mounted Path
	nodesList := &corev1.NodeList{}
	err := r.List(ctx, nodesList)
	if err != nil {
		fmt.Print("Error in Getting Nodes List")
		return false
	}

	for _, node := range nodesList.Items {
		command := "oc debug node/" + node.Name + " -- chroot /host rm -rf /mnt"
		_, out, _ := ExecuteCommand(command)
		fmt.Println(out)
	}
	fmt.Println("Mounted Paths Removed....")
	return true
}

//patchFinalizer patches finalizer in Resources
func patchFinalizer(rtype string, name string, namespace string) {
	_, out, _ := ExecuteCommand("kubectl patch " + rtype + " " + name + " -n " + namespace + " -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
	fmt.Println(out)
}

// localVolumeNSCleanUp performs cleanUp when namespace is in terminating state
func (r *CleanUpWatcher) localVolumeNSCleanUp(ctx context.Context, namespace string, resources []map[string]string, flag bool) bool {
	patchFinalizer("localvolumes.local.storage.openshift.io", "local-disk", namespace)
	if r.fetchRemovePV(ctx) && r.deleteMountedPath(ctx) {
		if flag {
			for index := range resources {
				resourceType := resources[index]["Type"]
				resourceName := resources[index]["Name"]
				resourceNamespace := resources[index]["Namespace"]
				if resourceType == "deployment" {
					patchFinalizer(resourceType, resourceName, resourceNamespace)
				}
			}
		}
		return true
	}
	return false
}
