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
	"context"
	"fmt"
	"reflect"
	"time"

	apps "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	practicev1alpha1 "github.com/ryanzhang-oss/deployment-watcher/api/v1alpha1"
)

// RyanReconciler reconciles a Ryan object
type RyanReconciler struct {
	client.Client
	record.EventRecorder
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=practice.shipa.io,resources=ryans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=practice.shipa.io,resources=ryans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=practice.shipa.io,resources=ryans/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;list;watch;update;patch
// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *RyanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	klog.InfoS("Reconcile ", klog.KRef(req.Namespace, req.Name))

	var helmResourceWatcher practicev1alpha1.Ryan
	if err := r.Get(ctx, req.NamespacedName, &helmResourceWatcher); err != nil {
		if apierrors.IsNotFound(err) {
			klog.Info("helmResourceWatcher is deleted")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	klog.InfoS("Get the helmResourceWatcher", "name", helmResourceWatcher.GetName(), "namespace", helmResourceWatcher.GetNamespace())

	if helmResourceWatcher.Spec.Kind == reflect.TypeOf(apps.Deployment{}).Name() &&
		helmResourceWatcher.Spec.APIVersion == apps.SchemeGroupVersion.String() {
		// check on the deployment
		var deploy apps.Deployment
		nsn := types.NamespacedName{Name: helmResourceWatcher.Spec.ResourceName, Namespace: helmResourceWatcher.Spec.ResourceNamespace}
		if err := r.Get(ctx, nsn, &deploy); err != nil {
			klog.ErrorS(err, "failed to get the deployment", "name", helmResourceWatcher.Spec.ResourceName, "namespace",
				helmResourceWatcher.Spec.ResourceNamespace)
			return ctrl.Result{}, err
		}

		annots := deploy.GetAnnotations()
		labels := deploy.GetLabels()
		if annots == nil || labels == nil ||
			len(annots["meta.helm.sh/release-name"]) == 0 ||
			labels["app.kubernetes.io/managed-by"] != "Helm" {
			err := fmt.Errorf("the workload is found but not managed by Helm")
			klog.ErrorS(err, "Found a name-matched workload but not managed by Helm", "name", helmResourceWatcher.Spec.ResourceName,
				"annotations", annots, "labels", labels)
			return ctrl.Result{}, err
		}
		klog.InfoS("Find the helmResource to watch", "name", nsn.Name, "namespace", nsn.Namespace)
		isOwner := false
		for _, ref := range deploy.GetOwnerReferences() {
			if ref.UID == helmResourceWatcher.GetUID() {
				isOwner = true
			}
		}
		if isOwner {
			ref := metav1.OwnerReference{
				APIVersion:         practicev1alpha1.GroupVersion.String(),
				Kind:               "Ryan",
				Name:               helmResourceWatcher.GetName(),
				UID:                helmResourceWatcher.GetUID(),
				BlockOwnerDeletion: pointer.BoolPtr(false),
				Controller:         pointer.BoolPtr(false),
			}
			deploy.SetOwnerReferences(append(deploy.GetOwnerReferences(), ref))
			if err := r.Update(ctx, &deploy); err != nil {
				klog.ErrorS(err, "Failed to claim the deployment", "name", nsn.Name, "namespace", nsn.Namespace)
				return ctrl.Result{}, err
			}
			klog.InfoS("Successfully claim the helmResource to watch", "name", nsn.Name, "namespace", nsn.Namespace)
		}
		helmResourceWatcher.Status.ReleaseName = annots["meta.helm.sh/release-name"]
		helmResourceWatcher.Status.AppName = labels["app.kubernetes.io/name"]
		helmResourceWatcher.Status.AppVersion = labels["app.kubernetes.io/version"]
		klog.InfoS("Find the helm app version", "name", nsn.Name, "namespace", nsn.Namespace)
		return ctrl.Result{}, r.Status().Update(ctx, &helmResourceWatcher)
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RyanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&practicev1alpha1.Ryan{}).
		// Owns(&apps.Deployment{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Watches(&source.Kind{Type: &apps.Deployment{}}, &handler.EnqueueRequestForOwner{
			OwnerType:    &practicev1alpha1.Ryan{},
			IsController: false,
		}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Complete(r)
}
