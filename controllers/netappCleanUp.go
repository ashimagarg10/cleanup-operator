package controllers

import "fmt"

// removeCRDs patches and deletes all trident crds
func removeCRDs(resources []map[string]string, flag bool) {
	crdNames := []string{"tridentbackends.trident.netapp.io", "tridentsnapshots.trident.netapp.io", "tridentstorageclasses.trident.netapp.io",
		"tridenttransactions.trident.netapp.io", "tridentvolumes.trident.netapp.io"}
	for index := range crdNames {
		crd := crdNames[index]
		patchFinalizer("crd", crd, "default")
		_, out, _ := ExecuteCommand("kubectl delete crd " + crd)
		fmt.Println(out)
	}

	_, out, _ := ExecuteCommand("kubectl patch crd/tridentversions.trident.netapp.io -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
	fmt.Println(out)
	_, out, _ = ExecuteCommand("kubectl patch crd/tridentversions.trident.netapp.io -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
	fmt.Println(out)
	_, out, _ = ExecuteCommand("kubectl delete crd tridentversions.trident.netapp.io")
	fmt.Println(out)

	_, out, _ = ExecuteCommand("kubectl patch crd/tridentnodes.trident.netapp.io -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
	fmt.Println(out)
	_, out, _ = ExecuteCommand("kubectl patch crd/tridentnodes.trident.netapp.io -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
	fmt.Println(out)
	_, out, _ = ExecuteCommand("kubectl delete crd tridentnodes.trident.netapp.io")
	fmt.Println(out)

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
}
