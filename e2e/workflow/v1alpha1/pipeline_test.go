package v1alpha1_test

import (
	"context"
	"testing"

	"github.com/milesbxf/puppeteer/e2e"
	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestSimpleBuildPipeline(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rig, err := e2e.NewTestRig()
	// defer rig.TearDown()
	g.Expect(err).NotTo(gomega.HaveOccurred(), "setting up test rig")
	t.Logf("Test rig set up with namespace %s", rig.Namespace)

	// Load and create a simple Pipeline Config

	objs, err := e2e.LoadResourcesFromTestData("simple_build_pipeline_config.yaml")
	g.Expect(err).NotTo(gomega.HaveOccurred(), "loading pipeline config yaml")

	pipelineConfig, ok := objs[0].(*corev1alpha1.PipelineConfig)
	pipelineConfig.ObjectMeta.Namespace = rig.Namespace
	g.Expect(ok).To(gomega.BeTrue(), "loading pipeline config obj")
	err = rig.K8s.Create(context.TODO(), pipelineConfig)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "creating pipeline config obj")
	t.Logf("Created pipeline object %s", pipelineConfig.Name)

	// Manually trigger a Pipeline

	objs, err = e2e.LoadResourcesFromTestData("simple_build_pipeline.yaml")
	g.Expect(err).NotTo(gomega.HaveOccurred(), "loading pipeline yaml")

	pipeline, ok := objs[0].(*corev1alpha1.Pipeline)
	g.Expect(ok).To(gomega.BeTrue(), "loading pipeline obj")

	pipeline.ObjectMeta.Namespace = rig.Namespace

	err = rig.K8s.Create(context.TODO(), pipeline)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "creating pipeline obj")

	g.Eventually(func() error {
		p := &corev1alpha1.Pipeline{}
		return rig.K8s.Get(context.Background(), types.NamespacedName{Name: pipeline.Name, Namespace: rig.Namespace}, p)
	}, "5s").Should(gomega.Succeed(), "creating pipeline obj")
	t.Logf("Created pipeline object %s", pipeline.Name)

	// Now the pipeline instance controller reconciles. It wants to associate the instance
	// with an Artifact matching the config provided. If there isn't already one matching the Source,
	// it creates one, without filling in the reference.

	g.Eventually(
		func() (*corev1alpha1.PipelineArtifact, error) {
			return getArtifactFromPipeline(rig.K8s, pipeline.Name, rig.Namespace)
		},
		"30s",
	).ShouldNot(gomega.BeNil(), "waiting for pipeline artifact")

	a, err := getArtifactFromPipeline(rig.K8s, pipeline.Name, rig.Namespace)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "getting artifact from pipeline")
	t.Logf("Got name of artifact for pipeline input: '%s'", a.Name)

	g.Eventually(
		func() (*corev1alpha1.StorageReference, error) {
			artifact := &corev1alpha1.Artifact{}
			err = rig.K8s.Get(context.Background(), types.NamespacedName{Name: a.Name, Namespace: rig.Namespace}, artifact)
			if err != nil {
				return nil, err
			}
			return artifact.Status.Reference, nil
		},
		"2m",
	).ShouldNot(gomega.BeNil(), "getting artifact reference")

	artifact := &corev1alpha1.Artifact{}
	err = rig.K8s.Get(context.Background(), types.NamespacedName{Name: a.Name, Namespace: rig.Namespace}, artifact)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "getting artifact")
	t.Logf("Got artifact for pipeline instance input: '%s'", artifact.Name)

	g.Eventually(
		func() (corev1alpha1.StagePhase, error) {
			stageInstance := &corev1alpha1.Stage{}
			err = rig.K8s.Get(context.Background(), types.NamespacedName{Name: pipeline.Name + "-build-1", Namespace: rig.Namespace}, stageInstance)
			if err != nil {
				return "", err
			}
			return stageInstance.Status.Phase, nil
		},
		"2m",
	).Should(gomega.Equal(corev1alpha1.StageInProgress), "waiting for pipeline stage in progress")
	// First stage is triggered (Stage)
	// Spins up job with configured image and shell script
	// Build sidecar pulls local storage down into a shared volume
	// Waits for attached job to finish
	// Searches for outputs
	//   In this case we have a Docker image - it gets saved as a tarball and put in local storage
	//   Another artifact is created

}

func getArtifactFromPipeline(c client.Client, name, namespace string) (*corev1alpha1.PipelineArtifact, error) {
	p := &corev1alpha1.Pipeline{}
	err := c.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, p)
	if err != nil {
		return nil, err
	}
	return p.Spec.Inputs["scm-upstream"].Artifact, err
}
