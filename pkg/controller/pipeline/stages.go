package pipeline

import (
	"context"
	"fmt"

	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var stageLog = logf.Log.WithName("pipelineinstance_controller_stages")

func (r *ReconcilePipeline) reconcilePipelineStageInstance(pipeline *corev1alpha1.Pipeline, stage *corev1alpha1.PipelineStage) (shouldProgressToNext bool, err error) {
	ordinal := 1
	stageInstance := &corev1alpha1.PipelineStageInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-%d", pipeline.Name, stage.Name, ordinal),
			Namespace: pipeline.Namespace,
		},
	}
	if err := controllerutil.SetControllerReference(pipeline, stageInstance, r.scheme); err != nil {
		return false, err
	}

	found := &corev1alpha1.PipelineStageInstance{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: stageInstance.Name, Namespace: stageInstance.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		stageLog.Info("creating pipelinestageinstance", "namespace", stageInstance.Namespace, "name", stageInstance.Name)
		err = r.Create(context.TODO(), stageInstance)
		return false, err
	} else if err != nil {
		return false, err
	}

	switch found.Status.Phase {
	case corev1alpha1.PipelineStageInstanceInProgress, corev1alpha1.PipelineStageInstanceError:
		return false, nil
	case corev1alpha1.PipelineStageInstanceComplete:
		return true, nil
	default:
		return false, nil
	}
}
