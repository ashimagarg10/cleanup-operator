/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	cleanupv1 "github.com/operator/cleanup-operator/api/v1"
)

// CleanUpOperatorReconciler reconciles a CleanUpOperator object
type CleanUpOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// // CleanUpWatcher reconciles a CleanUpOperator object
// type CleanUpWatcher struct {
// 	client.Client
// 	Log    logr.Logger
// 	Scheme *runtime.Scheme
// }

// var template = ""
// var namespace = ""
// var resources = make([]map[string]string, 1)

var finalizer_name = "custom/finalizer"

//+kubebuilder:rbac:groups=cleanup.ibm.com,resources=cleanupoperators,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=cleanup.ibm.com,resources=cleanupoperators/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=cleanup.ibm.com,resources=cleanupoperators/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CleanUpOperator object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *CleanUpOperatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("cleanupoperator", req.NamespacedName)

	instance := &cleanupv1.CleanUpOperator{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("CleanUpOperator resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object.
		log.Error(err, "Failed to get CleanUpOperator")
		return ctrl.Result{}, err
	}

	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(instance.GetFinalizers(), finalizer_name) {
			controllerutil.AddFinalizer(instance, finalizer_name)
			if err := r.Update(ctx, instance); err != nil {
				log.Error(err, "Error is adding custom finalizer in CustomResoure ", instance.Name)
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(instance.GetFinalizers(), finalizer_name) {
			// Custom finalizer is present, so perform cleanup

			template := instance.Spec.ResourceName
			namespace := instance.Spec.Namespace

			if template == "trident" {
				fmt.Println("NetApp Trident")

				fmt.Println("Getting Namespace")
				res := &corev1.Namespace{}
				err = r.Get(ctx, types.NamespacedName{Name: namespace}, res)
				if err != nil {
					log.Error(err, "Error is getting NetApp Trident Namespace ", namespace)
					return ctrl.Result{}, err
				}
				if !res.ObjectMeta.DeletionTimestamp.IsZero() {
					err = removeCRDs()
					if err != nil {
						// Failed to perform CleanUp
						return ctrl.Result{}, err
					}
				}
				log.Info("NetApp Tridente Template Cleaned Successfully!!!")
			}

			// remove custom finalizer from the resource and update it.
			controllerutil.RemoveFinalizer(instance, finalizer_name)
			if err := r.Update(ctx, instance); err != nil {
				log.Error(err, "Error is removing custom finalizer from CustomResoure ", instance.Name)
				return ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the resource is being deleted
		return ctrl.Result{}, nil
	}

	// resources = make([]map[string]string, 1)
	// template = instance.Spec.ResourceName
	// namespace = instance.Spec.Namespace
	// listOfResources := instance.Spec.Resources
	// for index := range listOfResources {
	// 	rmap := make(map[string]string)
	// 	rmap["Type"] = listOfResources[index].Type
	// 	rmap["Name"] = listOfResources[index].Name
	// 	rmap["Namespace"] = listOfResources[index].Namespace
	// 	resources = append(resources, rmap)
	// }
	// fmt.Println(template, namespace, resources)

	// // Check if the Namespace exists, if exists add custom finalizer
	// fmt.Println("Getting Namespace")
	// res := &corev1.Namespace{}
	// err = r.Get(ctx, types.NamespacedName{Name: namespace}, res)
	// if err != nil {
	// 	log.Error(err, "Error in Getting Namespace ", namespace)
	// 	return ctrl.Result{}, err
	// }
	// if !containsString(res.GetFinalizers(), finalizer_name) {
	// 	controllerutil.AddFinalizer(res, finalizer_name)
	// 	err = r.Update(ctx, res)
	// 	if err != nil {
	// 		log.Error(err, "Error is updating namespace ", namespace)
	// 		return ctrl.Result{}, err
	// 	}
	// }

	// for index := range resources {
	// 	resourceType := resources[index]["Type"]
	// 	resourceName := resources[index]["Name"]
	// 	resourceNamespace := resources[index]["Namespace"]

	// 	if resourceType == "deployment" {
	// 		fmt.Println("Getting Deployment")
	// 		res := &appsv1.Deployment{}
	// 		err = r.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: resourceNamespace}, res)
	// 		if err != nil {
	// 			log.Error(err, "Error in Getting Deployment ", resourceName)
	// 			return ctrl.Result{}, err
	// 		}
	// 		if !containsString(res.GetFinalizers(), finalizer_name) {
	// 			controllerutil.AddFinalizer(res, finalizer_name)
	// 			err = r.Update(ctx, res)
	// 			if err != nil {
	// 				log.Error(err, "Error is updating resource ", resourceName)
	// 				return ctrl.Result{}, err
	// 			}
	// 		}
	// 	} else if resourceType == "localvolume" {
	// 		fmt.Println("Getting Local-Volume")
	// 		res := &localv1.LocalVolume{}
	// 		err = r.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: resourceNamespace}, res)
	// 		if err != nil {
	// 			log.Error(err, "Error in Getting Local Volume ", resourceName)
	// 			return ctrl.Result{}, err
	// 		}
	// 		if !containsString(res.GetFinalizers(), finalizer_name) {
	// 			controllerutil.AddFinalizer(res, finalizer_name)
	// 			err = r.Update(ctx, res)
	// 			if err != nil {
	// 				log.Error(err, "Error is updating resource ", resourceName)
	// 				return ctrl.Result{}, err
	// 			}
	// 		}
	// 	} else if index != 0 {
	// 		fmt.Println("Resource ", resourceType, " is not being watched")
	// 	}
	// }

	return ctrl.Result{}, nil
}

// func (r *CleanUpWatcher) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
// 	fmt.Println("In watcher - ", req.Namespace)

// 	instance := &cleanupv1.CleanUpOperator{}
// 	err := r.Get(ctx, req.NamespacedName, instance)
// 	if err != nil {
// 		if errors.IsNotFound(err) {
// 			if template == "local-volume" {
// 				fmt.Println("Local Volume")

// 				fmt.Println("Getting Namespace")
// 				res := &corev1.Namespace{}
// 				err = r.Get(ctx, types.NamespacedName{Name: namespace}, res)
// 				if err != nil {
// 					fmt.Print("Error in Getting Namespace")
// 					return ctrl.Result{}, err
// 				}
// 				if !res.ObjectMeta.DeletionTimestamp.IsZero() {
// 					if containsString(res.GetFinalizers(), finalizer_name) && r.localVolumeNSCleanUp(ctx, namespace, resources, true) {
// 						fmt.Println("Custom finalizer Present")
// 						_, out, _ := ExecuteCommand("kubectl patch ns " + namespace + " -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
// 						fmt.Println(out)
// 						fmt.Println("Local Volume Template Cleaned Successfully!!!")
// 					} // else if r.localVolumeNSCleanUp(ctx, namespace, resources, false) {
// 					// 	fmt.Println("Local Volume Template Cleaned Successfully!!!")
// 					// }
// 				}
// 			} else if template == "trident" {
// 				fmt.Println("NetApp Trident")

// 				fmt.Println("Getting Namespace")
// 				res := &corev1.Namespace{}
// 				err = r.Get(ctx, types.NamespacedName{Name: namespace}, res)
// 				if err != nil {
// 					fmt.Print("Error in Getting Namespace")
// 					return ctrl.Result{}, err
// 				}
// 				if !res.ObjectMeta.DeletionTimestamp.IsZero() {
// 					if containsString(res.GetFinalizers(), finalizer_name) {
// 						fmt.Println("Custom finalizer Present")
// 						removeCRDs(resources, true)
// 						_, out, _ := ExecuteCommand("kubectl patch ns " + namespace + " -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
// 						fmt.Println(out)
// 						fmt.Println("NetApp Tridente Template Cleaned Successfully!!!")
// 					} // else {
// 					// 	removeCRDs(resources, false)
// 					// 	fmt.Println("NetApp Tridente Template Cleaned Successfully!!!")
// 					// }
// 				}
// 			}

// 			_, out, _ := ExecuteCommand("tridentctl version -n trident")
// 			fmt.Println(out)

// 			resources = make([]map[string]string, 1)
// 		}
// 	}

// 	return ctrl.Result{}, nil
// }

// SetupWithManager sets up the controller with the Manager.
func (r *CleanUpOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cleanupv1.CleanUpOperator{}).
		Complete(r)
}

// // SetupWithManager sets up the controller with the Manager.
// func (r *CleanUpWatcher) SetupWithManager(mgr ctrl.Manager) error {
// 	return ctrl.NewControllerManagedBy(mgr).
// 		For(&cleanupv1.CleanUpOperator{}).
// 		Complete(r)
// }

// Helper function to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// ExecuteCommand to execute shell commands
func ExecuteCommand(command string) (int, string, error) {
	fmt.Println("in ExecuteCommand")
	var cmd *exec.Cmd
	var cmdErr bytes.Buffer
	var cmdOut bytes.Buffer
	cmdErr.Reset()
	cmdOut.Reset()

	cmd = exec.Command("bash", "-c", command)
	cmd.Stderr = &cmdErr
	cmd.Stdout = &cmdOut
	err := cmd.Run()

	var waitStatus syscall.WaitStatus

	errStr := strings.TrimSpace(cmdErr.String())
	outStr := strings.TrimSpace(cmdOut.String())

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus = exitError.Sys().(syscall.WaitStatus)
		}
		if errStr != "" {
			fmt.Println(command)
			fmt.Println(errStr)
		}
	} else {
		waitStatus = cmd.ProcessState.Sys().(syscall.WaitStatus)
	}
	if waitStatus.ExitStatus() == -1 {
		fmt.Print(time.Now().String() + " Timed out " + command)
	}
	return waitStatus.ExitStatus(), outStr, err
}

// //patchFinalizer patches finalizer in Resources
// func patchFinalizer(rtype string, name string, namespace string) {
// 	_, out, _ := ExecuteCommand("kubectl patch " + rtype + " " + name + " -n " + namespace + " -p '{\"metadata\":{\"finalizers\":[]}}' --type=merge")
// 	fmt.Println(out)
// }
