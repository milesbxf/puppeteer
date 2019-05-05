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

package artifact

import (
	"context"
	"encoding/json"
	"fmt"

	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	pluginsv1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/plugins/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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

// Add creates a new Artifact Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileArtifact{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("artifact-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Artifact
	err = c.Watch(&source.Kind{Type: &corev1alpha1.Artifact{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileArtifact{}

// ReconcileArtifact reconciles a Artifact object
type ReconcileArtifact struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Artifact object and makes changes based on the state read
// and what is in the Artifact.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=artifacts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=artifacts/status,verbs=get;update;patch
func (r *ReconcileArtifact) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logParams := []interface{}{"artifact_name", request.NamespacedName.String()}

	// Fetch the Artifact instance
	instance := &corev1alpha1.Artifact{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			log.V(2).Info("Artifact not found, not reconciling further", logParams...)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Error reconciling Artifact", logParams...)
		return reconcile.Result{}, err
	}

	// Make sure an appropriate phase is set

	var phase corev1alpha1.ArtifactPhase
	switch {
	case instance.Spec.Source.Type == "":
		log.V(1).Info("Artifact has missing source type", logParams...)
		phase = corev1alpha1.InvalidArtifact
	case instance.Spec.Source.Type != "" && instance.Spec.Reference == "":
		phase = corev1alpha1.UnresolvedArtifact
	case instance.Spec.Source.Type != "" && instance.Spec.Reference != "":
		phase = corev1alpha1.ResolvedArtifact
	default:
		phase = corev1alpha1.InvalidArtifact
	}

	if phase != instance.Status.Phase {
		instance.Status.Phase = phase
		err = r.Client.Update(context.Background(), instance)
		if err != nil {
			log.Error(err, "Error updating Artifact status", logParams...)
			return reconcile.Result{}, err
		}
	}

	switch instance.Status.Phase {
	case corev1alpha1.UnresolvedArtifact:
		// Artifact has a source, but is not resolved - we'll trigger resolution based on the type
		log.V(2).Info("Resolving artifact", logParams...)

		// TODO: make this dynamically look up a plugin
		switch instance.Spec.Source.Type {
		case "git":

			// Check if a GitArtifactResolution already exists
			err := r.Client.Get(context.Background(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, &pluginsv1alpha1.GitArtifactResolution{})
			switch {
			case err == nil:
				log.V(3).Info("GitArtifactResolution already exists, not creating", logParams...)
				return reconcile.Result{}, nil
			case !apierrors.IsNotFound(err):
				log.Error(err, "Couldn't look up GitResolutionArtifact", logParams...)
				return reconcile.Result{}, err
			}

			// Parse config into a spec
			gitSpec := &pluginsv1alpha1.GitArtifactResolutionSpec{}
			err = json.Unmarshal([]byte(instance.Spec.Source.Config), gitSpec)
			if err != nil {
				log.Error(err, "Couldn't parse Artifact source config", append(logParams, "artifact_source_config", instance.Spec.Source.Type)...)
				return reconcile.Result{}, err
			}

			gitResolution := &pluginsv1alpha1.GitArtifactResolution{
				ObjectMeta: metav1.ObjectMeta{
					Name:      instance.Name,
					Namespace: instance.Namespace,
				},
				Spec: gitSpec,
			}
			err = r.Client.Create(context.Background(), gitResolution)
			if err != nil {
				log.Error(err, "Couldn't create GitResolutionArtifact", logParams...)
				return reconcile.Result{}, err
			}

		default:
			err := fmt.Errorf("unknown artifact source type %s", instance.Spec.Source.Type)
			log.Error(err, "Couldn't resolve Artifact", logParams...)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
