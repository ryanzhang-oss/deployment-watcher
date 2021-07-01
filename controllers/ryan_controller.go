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
	"strings"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	practicev1alpha1 "github.com/ryanzhang-oss/deployment-watcher/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const helmSecretPrefix = "sh.helm.release.v1."

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
	klog.InfoS("Reconcile get a secret", "name", req.Name, "namespace", req.Namespace)
	if !strings.HasPrefix(req.Name, helmSecretPrefix) {
		klog.InfoS("get a secret not created by helm", "name", req.Name, "namespace", req.Namespace)
		return ctrl.Result{}, nil
	}
	releaseName := strings.Split(strings.TrimPrefix(req.Name, helmSecretPrefix), ".")[0]
	var helmResourceWatcher practicev1alpha1.Ryan
	nsn := types.NamespacedName{Name: releaseName, Namespace: req.Namespace}
	if err := r.Get(ctx, nsn, &helmResourceWatcher); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		klog.ErrorS(err, "failed to get the helmResourceWatcher", "name", nsn.Name, "namespace", nsn.Namespace)
		return ctrl.Result{}, err
	}
	// this is the release we care about
	klog.InfoS("Reconcile get a secret we want to watch", "releaseName", releaseName,
		"resource name", helmResourceWatcher.Spec.ResourceName, "apiVersion", helmResourceWatcher.Spec.APIVersion,
		"kind", helmResourceWatcher.Spec.Kind)
	var helmObj unstructured.Unstructured
	helmObj.SetAPIVersion(helmResourceWatcher.Spec.APIVersion)
	helmObj.SetKind(helmResourceWatcher.Spec.Kind)
	nsn = types.NamespacedName{Name: helmResourceWatcher.Spec.ResourceName, Namespace: req.Namespace}
	if err := r.Get(ctx, nsn, &helmObj); err != nil {
		klog.ErrorS(err, "failed to get the helm resource", "name", helmResourceWatcher.Spec.ResourceName, "namespace",
			req.Namespace)
		return ctrl.Result{}, err
	}
	klog.InfoS("Reconcile get a helm resource", "name", helmResourceWatcher.Spec.ResourceName, "namespace",
		req.Namespace)
	annots := helmObj.GetAnnotations()
	labels := helmObj.GetLabels()
	if annots == nil || labels == nil ||
		len(annots["meta.helm.sh/release-name"]) == 0 ||
		labels["app.kubernetes.io/managed-by"] != "Helm" {
		err := fmt.Errorf("the workload is found but not managed by Helm")
		klog.ErrorS(err, "Found a name-matched workload but not managed by Helm", "name", helmResourceWatcher.Spec.ResourceName,
			"annotations", annots, "labels", labels)
		return ctrl.Result{}, err
	}
	klog.InfoS("Find the helmResource to watch", "name", nsn.Name, "namespace", nsn.Namespace)
	helmResourceWatcher.Status.ReleaseName = annots["meta.helm.sh/release-name"]
	helmResourceWatcher.Status.AppName = labels["app.kubernetes.io/name"]
	helmResourceWatcher.Status.AppVersion = labels["app.kubernetes.io/version"]
	klog.InfoS("Find the helm app version", "name", nsn.Name, "namespace", nsn.Namespace)
	return ctrl.Result{}, r.Status().Update(ctx, &helmResourceWatcher)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RyanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Secret{}).
		Complete(r)
}
