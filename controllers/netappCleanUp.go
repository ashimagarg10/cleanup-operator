package controllers

import (
	"context"
	"fmt"

	apiextenstionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// removeCRDs patches and deletes all trident crds
// func removeCRDs(resources []map[string]string, flag bool) error {
// func removeCRDs() error {
// 	crdNames := []string{"tridentbackends.trident.netapp.io", "tridentsnapshots.trident.netapp.io", "tridentstorageclasses.trident.netapp.io",
// 		"tridenttransactions.trident.netapp.io", "tridentvolumes.trident.netapp.io", "tridentversions.trident.netapp.io", "tridentnodes.trident.netapp.io"}
// 	for index := range crdNames {
// 		crd := crdNames[index]
// 		// 		patchFinalizer("crd", crd, "default")
// 		_, out, err := ExecuteCommand("kubectl patch crd/" + crd + " -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
// 		if err != nil {
// 			fmt.Println("Error in patching crd: ", crd)
// 			return err
// 		}
// 		fmt.Println(out)
// 		_, out, err = ExecuteCommand("kubectl delete crd " + crd)
// 		if err != nil {
// 			fmt.Println("Error in deleting crd: ", crd)
// 			return err
// 		}
// 		fmt.Println(out)
// 	}

// 	// 	_, out, _ := ExecuteCommand("kubectl patch crd/tridentversions.trident.netapp.io -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
// 	// 	fmt.Println(out)
// 	// 	// _, out, _ = ExecuteCommand("kubectl patch crd/tridentversions.trident.netapp.io -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
// 	// 	// fmt.Println(out)
// 	// 	_, out, _ = ExecuteCommand("kubectl delete crd tridentversions.trident.netapp.io")
// 	// 	fmt.Println(out)

// 	// 	_, out, _ = ExecuteCommand("kubectl patch crd/tridentnodes.trident.netapp.io -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
// 	// 	fmt.Println(out)
// 	// 	// _, out, _ = ExecuteCommand("kubectl patch crd/tridentnodes.trident.netapp.io -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
// 	// 	// fmt.Println(out)
// 	// 	_, out, _ = ExecuteCommand("kubectl delete crd tridentnodes.trident.netapp.io")
// 	// 	fmt.Println(out)

// 	// 	if flag {
// 	// 		for index := range resources {
// 	// 			resourceType := resources[index]["Type"]
// 	// 			resourceName := resources[index]["Name"]
// 	// 			resourceNamespace := resources[index]["Namespace"]

// 	// 			if resourceType == "deployment" {
// 	// 				patchFinalizer(resourceType, resourceName, resourceNamespace)
// 	// 			}
// 	// 		}
// 	// 	}
// 	return nil
// }

func (cr CleanUpOperatorReconciler) removeCRDs(ctx context.Context) error {

	// type patchFinalizer struct {
	// 	MetaData map[string]string `json:"metadata"`
	// }

	// var payload = []patchFinalizer{{
	// 	MetaData: map[string]string{"finalizers": ""},
	// }}

	// var clientSet *kubernetes.Clientset

	crdNames := []string{"tridentbackends.trident.netapp.io", "tridentsnapshots.trident.netapp.io", "tridentstorageclasses.trident.netapp.io",
		"tridenttransactions.trident.netapp.io", "tridentvolumes.trident.netapp.io", "tridentversions.trident.netapp.io", "tridentnodes.trident.netapp.io"}
	for index := range crdNames {
		crd := crdNames[index]

		CRD := &apiextenstionsv1.CustomResourceDefinition{}
		err := cr.Get(ctx, types.NamespacedName{Name: crd}, CRD)
		if err != nil {
			fmt.Println("Error in getting crd: ", crd)
			return err
		}
		fmt.Println(CRD.Name)

		finalizersInCRD := CRD.ObjectMeta.GetObjectMeta().GetFinalizers()
		fmt.Println(finalizersInCRD)

		for index := range finalizersInCRD {
			controllerutil.RemoveFinalizer(CRD, finalizersInCRD[index])
			if err := cr.Update(ctx, CRD); err != nil {
				fmt.Println(err, "Error is removing finalizers from CustomResoure ", CRD.Name)
				return err
			}
		}

		err = cr.Delete(ctx, CRD)
		if err != nil {
			fmt.Println(err, "Error is deleting CustomResoure ", CRD.Name)
			return err
		}

		// payloadBytes, _ := json.Marshal(payload)

		// err = cr.Client.Patch(ctx, CRD, client.Merge, &client.PatchOptions{})

		// fmt.Println(crd)

		// res, err := clientSet.CoreV1().
		// 	ReplicationControllers("default").Patch(ctx, crd, types.JSONPatchType, payloadBytes, v1.PatchOptions{})

		// fmt.Println(res)
		// fmt.Println()
		// fmt.Println(err)

		// _, out, err := ExecuteCommand("kubectl patch crd/" + crd + " -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
		// if err != nil {
		// 	fmt.Println("Error in patching crd: ", crd)
		// 	return err
		// }
		// fmt.Println(out)
		// _, out, err = ExecuteCommand("kubectl delete crd " + crd)
		// if err != nil {
		// 	fmt.Println("Error in deleting crd: ", crd)
		// 	return err
		// }
		// fmt.Println(out)
	}

	return nil
}
