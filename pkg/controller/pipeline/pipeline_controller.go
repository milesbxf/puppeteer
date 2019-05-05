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

package pipeline

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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("pipeline-controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Pipeline Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcilePipeline{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("pipeline-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Pipeline
	err = c.Watch(&source.Kind{Type: &corev1alpha1.Pipeline{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for artifacts this pipeline owns
	err = c.Watch(&source.Kind{Type: &corev1alpha1.Artifact{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &corev1alpha1.Pipeline{},
	})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcilePipeline{}

// ReconcilePipeline reconciles a Pipeline object
type ReconcilePipeline struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Pipeline object and makes changes based on the state read
// and what is in the Pipeline.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=pipelines,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.puppeteer.milesbryant.co.uk,resources=pipelines/status,verbs=get;update;patch
func (r *ReconcilePipeline) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logParams := []interface{}{"pipeline_instance_name", request.NamespacedName.String()}

	log.V(2).Info("Reconciling Pipeline object", logParams...)

	// Fetch the Pipeline instance
	instance := &corev1alpha1.Pipeline{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			log.V(2).Info("Pipeline not found, not reconciling further", logParams...)
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Error reconciling Pipeline ", logParams...)
		return reconcile.Result{}, err
	}

	// Fetch the associated pipeline config
	pipelineConfig := &corev1alpha1.PipelineConfig{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.PipelineName, Namespace: instance.Namespace}, pipelineConfig)
	if err != nil {
		if errors.IsNotFound(err) {
			err = fmt.Errorf("pipeline config %s referenced in pipelineinstance %s doesn't exist", instance.Spec.PipelineName, instance.Name)
			log.Error(err, "Error reconciling Pipeline ", logParams...)
			return reconcile.Result{}, err
		}
		log.Error(err, "Error reconciling Pipeline ", logParams...)
		return reconcile.Result{}, err
	}

	for key, input := range instance.Spec.Inputs {
		innerLogParams := append(logParams, "pipeline_instance_input", key)

		// Look up corresponding pipeline input
		pipelineInputConfig, ok := pipelineConfig.Spec.Inputs[key]
		if !ok {
			err = fmt.Errorf("input %s referenced in pipelineinstance %s doesn't exist in pipeline %s", key, instance.Name, instance.Spec.PipelineName)
			log.Error(err, "Error reconciling Pipeline ", innerLogParams...)
			return reconcile.Result{}, err
		}

		if input.Artifact == nil {
			err := r.createArtifactForPipelineInput(instance, key, input, &pipelineInputConfig, innerLogParams)
			return reconcile.Result{}, err
		}

		artifact := &corev1alpha1.Artifact{}
		err = r.Client.Get(context.Background(), types.NamespacedName{Name: input.Artifact.Name, Namespace: instance.Namespace}, artifact)
		if err != nil {
			log.Error(err, "lookup artifact reference", append(innerLogParams, "artifact_name", input.Artifact.Name)...)
			return reconcile.Result{}, err
		}

		if artifact.Status.Phase == corev1alpha1.ResolvedArtifact {
			// Make sure a pipeline stage instance exists for each stage in sequence
			for _, stage := range pipelineConfig.Spec.Workflow.Stages {
				progressNextStage, err := r.reconcilePipelineStageInstance(instance, &stage)
				if err != nil {
					log.Error(err, "reconciling stage instance", append(innerLogParams, "artifact_name", input.Artifact.Name, "stage_name", stage.Name)...)
					return reconcile.Result{}, err
				}
				if !progressNextStage {
					break
				}
			}
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcilePipeline) createArtifactForPipelineInput(instance *corev1alpha1.Pipeline, inputName string, pipelineInput *corev1alpha1.PipelineInput, pipelineInputConfig *corev1alpha1.PipelineInputConfig, logParams []interface{}) error {
	// No artifact attached, we need to find or create one

	switch pipelineInput.Type {
	case "git":
		repoConfig, err := pluginsv1alpha1.RepoConfigFromJSON(pipelineInputConfig.Config)
		if err != nil {
			log.Error(err, "Error parsing pipeline input config", append(logParams, "pipeline_input_config", pipelineInputConfig.Config)...)
			return err
		}

		triggerConfig, err := pluginsv1alpha1.GitPipelineInputConfigFromJSON(pipelineInput.Config)
		if err != nil {
			log.Error(err, "Error parsing pipeline instance input config", append(logParams, "pipeline_instance_input_config", pipelineInput.Config)...)
			return err
		}

		gitSpec := &pluginsv1alpha1.GitArtifactResolutionSpec{
			RepositoryURL: repoConfig.Repository,
			CommitSHA:     triggerConfig.Commit,
		}

		artifactConfig := gitSpec.ToJSON()

		configHash := fmt.Sprintf("%x", sha256.Sum224([]byte(artifactConfig)))
		artifactName := fmt.Sprintf("%s-%s", pipelineInput.Type, configHash)
		logParams := append(logParams, "pipeline_instance_input", inputName, "artifact_name", artifactName)

		log.V(2).Info("No artifact attached to Pipeline, looking for existing one", logParams...)

		artifact := &corev1alpha1.Artifact{}
		err = r.Client.Get(context.Background(), types.NamespacedName{
			Name:      artifactName,
			Namespace: instance.Namespace,
		}, artifact)

		if err != nil && !apierrors.IsNotFound(err) {
			log.Error(err, "Error looking up Artifact for Pipeline", logParams...)
			return err
		}

		if apierrors.IsNotFound(err) {
			log.V(2).Info("Didn't find Artifact for Pipeline, creating it", logParams...)

			source := corev1alpha1.ArtifactSource{
				Type:   pipelineInput.Type,
				Config: artifactConfig,
			}

			artifact = &corev1alpha1.Artifact{
				ObjectMeta: metav1.ObjectMeta{
					Name:      artifactName,
					Namespace: instance.Namespace,
					Labels: map[string]string{
						"v1alpha1.core.puppeteer.milesbryant.co.uk/source-type":        pipelineInput.Type,
						"v1alpha1.core.puppeteer.milesbryant.co.uk/source-config-hash": configHash,
					}},
				Spec: corev1alpha1.ArtifactSpec{
					Source: source,
				},
			}
			if err := controllerutil.SetControllerReference(instance, artifact, r.scheme); err != nil {
				return err
			}
			err := r.Client.Create(context.Background(), artifact)
			if err != nil {
				log.Error(err, "Error creating Artifact for Pipeline", logParams...)
				return err
			}

		}
		instance.Spec.Inputs[inputName].Artifact = &corev1alpha1.PipelineArtifact{
			Name: artifact.Name,
		}

		log.V(2).Info("Updating Pipeline Input reference to Artifact", logParams...)
		err = r.Client.Update(context.Background(), instance)
		if err != nil {
			log.Error(err, "Error updating Pipeline with Artifact reference", logParams...)
			return err
		}
	default:
		err := fmt.Errorf("unknown pipeline input type %s", pipelineInput.Type)
		log.Error(err, "Couldn't create Artifact for pipelineinstance", logParams...)
		return err
	}

	return nil
}
