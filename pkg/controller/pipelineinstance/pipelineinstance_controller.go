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

package pipelineinstance

import (
	"context"
	"crypto/sha256"
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

var log = logf.Log.WithName("pipelineinstance_controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new PipelineInstance Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePipelineInstance{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pipelineinstance-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to PipelineInstance
	err = c.Watch(&source.Kind{Type: &corev1alpha1.PipelineInstance{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcilePipelineInstance{}

// ReconcilePipelineInstance reconciles a PipelineInstance object
type ReconcilePipelineInstance struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a PipelineInstance object and makes changes based on the state read
// and what is in the PipelineInstance.Spec
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=pipelineinstances,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=pipelineinstances/status,verbs=get;update;patch
func (r *ReconcilePipelineInstance) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	logParams := []interface{}{"pipeline_instance_name", request.NamespacedName.String()}

	log.V(2).Info("Reconciling PipelineInstance object", logParams...)

	// Fetch the PipelineInstance instance
	instance := &corev1alpha1.PipelineInstance{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			log.V(2).Info("PipelineInstance not found, not reconciling further", logParams...)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Error reconciling PipelineInstance ", logParams...)
		return reconcile.Result{}, err
	}

	// Fetch the associated pipeline
	pipeline := &corev1alpha1.Pipeline{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.PipelineName, Namespace: instance.Namespace}, pipeline)
	if err != nil {
		if errors.IsNotFound(err) {
			err = fmt.Errorf("pipeline %s referenced in pipelineinstance %s doesn't exist", instance.Spec.PipelineName, instance.Name)
			log.Error(err, "Error reconciling PipelineInstance ", logParams...)
			return reconcile.Result{}, err
		}
		log.Error(err, "Error reconciling PipelineInstance ", logParams...)
		return reconcile.Result{}, err
	}

	for key, input := range instance.Spec.Inputs {
		innerLogParams := append(logParams, "pipeline_instance_input", key)

		// Look up corresponding pipeline input
		pipelineInput, ok := pipeline.Spec.Inputs[key]
		if !ok {
			err = fmt.Errorf("input %s referenced in pipelineinstance %s doesn't exist in pipeline %s", key, instance.Name, instance.Spec.PipelineName)
			log.Error(err, "Error reconciling PipelineInstance ", innerLogParams...)
			return reconcile.Result{}, err
		}

		if input.Artifact == nil {
			// No artifact attached, we need to find or create one

			switch input.Type {
			case "git":
				pipelineInputConfig, err := pluginsv1alpha1.GitPipelineInputConfigFromJSON(pipelineInput.Config)
				if err != nil {
					log.Error(err, "Error parsing pipeline input config", append(innerLogParams, "pipeline_input_config", pipelineInput.Config)...)
					return reconcile.Result{}, err
				}

				pipelineInstanceInputConfig, err := pluginsv1alpha1.GitPipelineInstanceInputConfigFromJSON(input.Config)
				if err != nil {
					log.Error(err, "Error parsing pipeline instance input config", append(innerLogParams, "pipeline_instance_input_config", input.Config)...)
					return reconcile.Result{}, err
				}

				gitSpec := &pluginsv1alpha1.GitArtifactResolutionSpec{
					RepositoryURL: pipelineInputConfig.Repository,
					CommitSHA:     pipelineInstanceInputConfig.Commit,
				}

				artifactConfig := gitSpec.ToJSON()

				configHash := fmt.Sprintf("%x", sha256.Sum224([]byte(artifactConfig)))
				artifactName := fmt.Sprintf("%s-%s", input.Type, configHash)
				innerLogParams := append(logParams, "pipeline_instance_input", key, "artifact_name", artifactName)

				log.V(2).Info("No artifact attached to PipelineInstance, looking for existing one", innerLogParams...)

				artifact := &corev1alpha1.Artifact{}
				err = r.Client.Get(context.Background(), types.NamespacedName{
					Name:      artifactName,
					Namespace: instance.Namespace,
				}, artifact)

				if err != nil && !apierrors.IsNotFound(err) {
					log.Error(err, "Error looking up Artifact for PipelineInstance", innerLogParams...)
					return reconcile.Result{}, err
				}

				if apierrors.IsNotFound(err) {
					log.V(2).Info("Didn't find Artifact for PipelineInstance, creating it", innerLogParams...)

					source := corev1alpha1.ArtifactSource{
						Type:   input.Type,
						Config: artifactConfig,
					}

					artifact = &corev1alpha1.Artifact{
						ObjectMeta: metav1.ObjectMeta{
							Name:      artifactName,
							Namespace: instance.Namespace,
							Labels: map[string]string{
								"v1alpha1.core.puppeteer.milesbryant.co.uk/source-type":        input.Type,
								"v1alpha1.core.puppeteer.milesbryant.co.uk/source-config-hash": configHash,
							}},
						Spec: corev1alpha1.ArtifactSpec{
							Source: source,
						},
					}
					err := r.Client.Create(context.Background(), artifact)
					if err != nil {
						log.Error(err, "Error creating Artifact for PipelineInstance", innerLogParams...)
						return reconcile.Result{}, err
					}

				}
				instance.Spec.Inputs[key].Artifact = &corev1alpha1.PipelineInstanceArtifact{}

				log.V(2).Info("Updating PipelineInstance Input reference to Artifact", innerLogParams...)
				err = r.Client.Update(context.Background(), instance)
				if err != nil {
					log.Error(err, "Error updating PipelineInstance with Artifact reference", innerLogParams...)
					return reconcile.Result{}, err
				}
			default:
				err := fmt.Errorf("unknown pipeline input type %s", input.Type)
				log.Error(err, "Couldn't create Artifact for pipelineinstance", logParams...)
				return reconcile.Result{}, err
			}

			return reconcile.Result{}, nil
		}
	}

	return reconcile.Result{}, nil
}