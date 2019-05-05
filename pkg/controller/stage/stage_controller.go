/*
Copyright 2019 Miles Bryant.

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

package stage

import (
	"context"

	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller")

// Add creates a new Stage Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileStage{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("stage-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Stage
	err = c.Watch(&source.Kind{Type: &corev1alpha1.Stage{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileStage{}

// ReconcileStage reconciles a Stage object
type ReconcileStage struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Stage object and makes changes based on the state read
// and what is in the Stage.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=stages,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=stages/status,verbs=get;update;patch
func (r *ReconcileStage) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logParams := []interface{}{"name", request.NamespacedName.String()}
	// Fetch the Stage instance
	instance := &corev1alpha1.Stage{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	phase := corev1alpha1.StageInProgress
	if phase != instance.Status.Phase {
		instance.Status.Phase = phase
		err = r.Client.Update(context.Background(), instance)
		if err != nil {
			log.Error(err, "Error updating stage status", logParams...)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
