/*
Copyright 2022.

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

package deployments

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/yardbirdsax/k8s-controller/apis/deployments/v1alpha1"
	deploymentsv1alpha1 "github.com/yardbirdsax/k8s-controller/apis/deployments/v1alpha1"
)

// WebAPIReconciler reconciles a WebAPI object
type WebAPIReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=deployments.k8s-controller.feiermanfamily.com,resources=webapis,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=deployments.k8s-controller.feiermanfamily.com,resources=webapis/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=deployments.k8s-controller.feiermanfamily.com,resources=webapis/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WebAPI object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *WebAPIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	webAPI := &v1alpha1.WebAPI{}
	err := r.Get(ctx, req.NamespacedName, webAPI)
	if err != nil {
		log.Error(err, "error retrieving webapi resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("webapi retrieved", "object", webAPI)

	newWebAPI := webAPI.DeepCopy()
	label(newWebAPI)
	err = r.Patch(ctx, newWebAPI, client.MergeFrom(webAPI))
	if err != nil {
		log.Error(err, "error patching webapi object")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebAPIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&deploymentsv1alpha1.WebAPI{}).
		Complete(r)
}

func label(webapi *v1alpha1.WebAPI) {
	if webapi.Labels == nil {
		webapi.Labels = map[string]string{}
	}
	webapi.Labels["labeled"] = "true"
}
