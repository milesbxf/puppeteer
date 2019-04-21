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

package gitsource

import (
	"context"
	"fmt"
	"reflect"

	pluginsv1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/plugins/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/util/pointer"
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

// Add creates a new GitSource Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGitSource{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gitsource-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to GitSource
	err = c.Watch(&source.Kind{Type: &pluginsv1alpha1.GitSource{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by GitSource - change this for objects you create
	err = c.Watch(&source.Kind{Type: &batchv1beta1.CronJob{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &pluginsv1alpha1.GitSource{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileGitSource{}

// ReconcileGitSource reconciles a GitSource object
type ReconcileGitSource struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a GitSource object and makes changes based on the state read
// and what is in the GitSource.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=cronjobs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=plugins.puppeteer.milesbryant.co.uk,resources=gitsources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=plugins.puppeteer.milesbryant.co.uk,resources=gitsources/status,verbs=get;update;patch
func (r *ReconcileGitSource) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the GitSource instance
	instance := &pluginsv1alpha1.GitSource{}
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

	cronjob := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-git-poller",
			Namespace: instance.Namespace,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:                   fmt.Sprintf("*/%d * * * *", *instance.Spec.Poll.IntervalMinutes),
			StartingDeadlineSeconds:    pointer.Int64Ptr(int64(*instance.Spec.Poll.IntervalMinutes * int32(60))),
			ConcurrencyPolicy:          batchv1beta1.ForbidConcurrent,
			SuccessfulJobsHistoryLimit: pointer.Int32Ptr(100),
			FailedJobsHistoryLimit:     pointer.Int32Ptr(100),
			JobTemplate:                getJobTemplate(instance),
		},
	}

	if err := controllerutil.SetControllerReference(instance, cronjob, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	found := &batchv1beta1.CronJob{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: cronjob.Name, Namespace: cronjob.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Cronjob", "namespace", cronjob.Namespace, "name", cronjob.Name)
		err = r.Create(context.TODO(), cronjob)
		return reconcile.Result{}, err
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(cronjob.Spec, found.Spec) {
		found.Spec = cronjob.Spec
		log.Info("Updating CronJob", "namespace", cronjob.Namespace, "name", cronjob.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func getJobTemplate(gitSource *pluginsv1alpha1.GitSource) batchv1beta1.JobTemplateSpec {
	return batchv1beta1.JobTemplateSpec{
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:  "git-poller",
							Image: "busybox",
							Args: []string{
								"/bin/sh",
								"-c",
								"date; echo Hello GitSource world!",
							},
						},
					},
				},
			},
		},
	}
}
