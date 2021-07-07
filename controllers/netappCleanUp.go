package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	corev1 "k8s.io/api/core/v1"
	apiextenstionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// removeCRDs patches and deletes all trident crds
func (cr *CleanUpOperatorReconciler) removeCRDs(ctx context.Context) error {
	//log := cr.Log.WithValues("cleanupoperator", "Removing NetApp Configuration")
	//defer logFunctionDuration(log, "removeCRDs", time.Now())
	starttime := time.Now()
	crdNames := []string{"tridentbackends.trident.netapp.io", "tridentsnapshots.trident.netapp.io", "tridentstorageclasses.trident.netapp.io",
		"tridenttransactions.trident.netapp.io", "tridentvolumes.trident.netapp.io", "tridentversions.trident.netapp.io", "tridentnodes.trident.netapp.io"}
	for _, crd := range crdNames {
		CRD := &apiextenstionsv1.CustomResourceDefinition{}
		err := cr.Get(ctx, types.NamespacedName{Name: crd}, CRD)
		if err != nil {
			if errors.IsNotFound(err) {
				fmt.Println("CRD not found: ", crd)
				continue
			}
			fmt.Println(err, "error in getting crd: ", crd)
			return err
		}

		CRD.SetFinalizers([]string{})
		if err := cr.Update(ctx, CRD); err != nil {
			fmt.Println(err, "Error is removing finalizers from CustomResoure ", CRD.Name)
			return err
		}

		err = cr.Delete(ctx, CRD)
		if err != nil {
			fmt.Println(err, "Error is deleting CustomResoure ", CRD.Name)
			return err
		}

		fmt.Println(CRD.Name)
	}
	duration := time.Since(starttime)
	fmt.Println("Time to complete", duration.Seconds())
	return nil
}

// patchCRs patches all tridentNodes and tridentVersions CRs
func (cr *CleanUpOperatorReconciler) patchCRs(ctx context.Context, namespace string) error {

	nodesList := &corev1.NodeList{}
	err := cr.List(ctx, nodesList)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Nodes List not found")
			return err
		}
		fmt.Println(err, "Error in getting Nodes List")
		return err
	}

	for _, node := range nodesList.Items {
		customResource := "tridentnodes.trident.netapp.io/" + node.Name
		command := "kubectl patch " + customResource + " -n " + namespace + " -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge"
		_, out, err := ExecuteCommand(command)
		if err != nil {
			if errors.IsNotFound(err) {
				fmt.Println("CR not found: tridentnodes.trident.netapp.io/", node.Name)
				return err
			}
			fmt.Println("Error in patching finalizer in CR: tridentnodes.trident.netapp.io/", node.Name)
			return err
		}
		fmt.Println(out)
	}

	ns := &corev1.Namespace{}
	err = cr.Get(ctx, types.NamespacedName{Name: namespace}, ns)
	if err != nil {
		fmt.Println("Namespace Not Found")
	} else {
		customResource := "tridentversions.trident.netapp.io/trident"
		command := "kubectl patch " + customResource + " -n " + namespace + " -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge"
		_, out, err := ExecuteCommand(command)
		if err != nil {
			if errors.IsNotFound(err) {
				fmt.Println("CR not found: tridentversions.trident.netapp.io/trident")
				return err
			}
			fmt.Println("Error in patching finalizer in CR: tridentversions.trident.netapp.io/trident")
			return err
		}
		fmt.Println(out)
	}
	return nil
}
