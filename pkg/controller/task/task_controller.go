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

package task

import (
	"context"

	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilpointer "k8s.io/kubernetes/pkg/util/pointer"
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

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Task Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileTask{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("task-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Task
	err = c.Watch(&source.Kind{Type: &corev1alpha1.Task{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileTask{}

// ReconcileTask reconciles a Task object
type ReconcileTask struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Task object and makes changes based on the state read
// and what is in the Task.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=tasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=tasks/status,verbs=get;update;patch
func (r *ReconcileTask) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Task instance
	task := &corev1alpha1.Task{}
	err := r.Get(context.TODO(), request.NamespacedName, task)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	if task.Spec.Config == nil {
		return reconcile.Result{}, nil
	}

	err = r.reconcileJob(task)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileTask) reconcileJob(task *corev1alpha1.Task) error {
	initVolume, initMount, taskEntrypoint, err := r.reconcileInitConfigMap(task)
	if err != nil {
		return err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      task.Name,
			Namespace: task.Namespace,
		},
		Spec: batchv1.JobSpec{
			Parallelism:  utilpointer.Int32Ptr(1),
			BackoffLimit: utilpointer.Int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						initVolume,
						{
							Name: "puppeteer-data",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  task.Name,
							Image: task.Spec.Config.Image,
							Command: []string{
								"bash",
								taskEntrypoint,
							},
							// TODO: what if this is relative? what if it's not specified?
							WorkingDir: task.Spec.Config.WorkingDir,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "puppeteer-data",
									MountPath: "/puppeteer-data",
								},
								initMount,
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(task, job, r.scheme); err != nil {
		return err
	}

	found := &batchv1.Job{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("creating job", "namespace", job.Namespace, "name", job.Name)
		err = r.Create(context.TODO(), job)
		if err != nil {
			return err
		}
		task.Status.Phase = corev1alpha1.TaskInProgress
		err = r.Update(context.Background(), task)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileTask) reconcileInitConfigMap(task *corev1alpha1.Task) (corev1.Volume, corev1.VolumeMount, string, error) {
	configmap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      task.Name + "-init",
			Namespace: task.Namespace,
		},
		Data: map[string]string{
			"entrypoint.sh": task.Spec.Config.Shell,
		},
	}
	if err := controllerutil.SetControllerReference(task, configmap, r.scheme); err != nil {
		return corev1.Volume{}, corev1.VolumeMount{}, "", err
	}

	found := &corev1.ConfigMap{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: configmap.Name, Namespace: configmap.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("creating init configmap", "namespace", configmap.Namespace, "name", configmap.Name)
		err = r.Create(context.TODO(), configmap)
		if err != nil {
			return corev1.Volume{}, corev1.VolumeMount{}, "", err
		}
	} else if err != nil {
		return corev1.Volume{}, corev1.VolumeMount{}, "", err
	}
	volume := corev1.Volume{
		Name: "puppeteer-init",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configmap.Name,
				},
			},
		},
	}

	mount := corev1.VolumeMount{
		Name:      volume.Name,
		ReadOnly:  true,
		MountPath: "/puppeteer-init",
	}

	return volume, mount, "/puppeteer-init/entrypoint.sh", nil
}
