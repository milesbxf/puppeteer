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

var stageLog = logf.Log.WithName("pipeline_controller_stages")

func (r *ReconcilePipeline) reconcileStage(pipeline *corev1alpha1.Pipeline, stageConfig *corev1alpha1.StageConfig) (shouldProgressToNext bool, err error) {
	ordinal := 1
	stage := &corev1alpha1.Stage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s-%d", pipeline.Name, stageConfig.Name, ordinal),
			Namespace: pipeline.Namespace,
		},
	}
	if err := controllerutil.SetControllerReference(pipeline, stage, r.scheme); err != nil {
		return false, err
	}

	found := &corev1alpha1.Stage{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: stage.Name, Namespace: stage.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		stageLog.Info("creating stage", "namespace", stage.Namespace, "name", stage.Name)
		err = r.Create(context.TODO(), stage)
		return false, err
	} else if err != nil {
		return false, err
	}

	switch found.Status.Phase {
	case corev1alpha1.StageInProgress, corev1alpha1.StageError:
		return false, nil
	case corev1alpha1.StageComplete:
		return true, nil
	default:
		return false, nil
	}
}
