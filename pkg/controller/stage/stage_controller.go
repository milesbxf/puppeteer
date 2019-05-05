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
	"fmt"

	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	// Watch for Task changes (e.g. status)
	err = c.Watch(&source.Kind{Type: &corev1alpha1.Task{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1alpha1.Stage{},
	})
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
	// Fetch the Stage
	stage := &corev1alpha1.Stage{}
	err := r.Get(context.TODO(), request.NamespacedName, stage)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	for name, taskConfig := range stage.Spec.Config.Tasks {
		err := r.reconcileTask(stage, name, &taskConfig)
		if err != nil {
			log.Error(err, "reconciling task", append(logParams, "task_name", name)...)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileStage) reconcileTask(stage *corev1alpha1.Stage, name string, config *corev1alpha1.TaskConfig) error {
	ordinal := 1

	task := &corev1alpha1.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-%d", stage.Name, name, ordinal),
			Namespace: stage.Namespace,
		},
		Spec: corev1alpha1.TaskSpec{
			Config: config,
		},
	}

	if err := controllerutil.SetControllerReference(stage, task, r.scheme); err != nil {
		return err
	}

	found := &corev1alpha1.Task{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: task.Name, Namespace: task.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {

		log.Info("creating task", "namespace", task.Namespace, "name", task.Name)
		err = r.Create(context.TODO(), task)
		return err
	} else if err != nil {
		return err
	}

	var phase corev1alpha1.StagePhase
	switch found.Status.Phase {
	case corev1alpha1.TaskInProgress:
		phase = corev1alpha1.StageInProgress
	case corev1alpha1.TaskComplete:
		phase = corev1alpha1.StageComplete
	default:
		return nil
	}

	if stage.Status.Phase != phase {
		log.Info("transition stage between phases", "name", stage.Name, "from", stage.Status.Phase, "to", phase)
		stage.Status.Phase = phase
		err = r.Update(context.Background(), stage)
		if err != nil {
			return err
		}
	}

	return nil
}
