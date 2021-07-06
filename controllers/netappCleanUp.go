package controllers

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

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
		fmt.Println(CRD.Name)
		/*finalizersInCRD := CRD.ObjectMeta.GetObjectMeta().GetFinalizers()
		fmt.Println(finalizersInCRD)
		for index := range finalizersInCRD {
			controllerutil.RemoveFinalizer(CRD, finalizersInCRD[index])
		}*/
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
	}
	duration := time.Since(starttime)
	fmt.Println("Time to complete", duration.Seconds())
	return nil
}
