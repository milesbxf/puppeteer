package v1alpha1_test

import (
	"context"
	"testing"

	"github.com/milesbxf/puppeteer/e2e"
	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

func TestSimpleBuildPipeline(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rig, err := e2e.NewTestRig()
	defer rig.TearDown()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Load and create a simple Pipeline

	objs, err := e2e.LoadResourcesFromTestData("simple_build_pipeline.yaml")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	pipeline, ok := objs[0].(*corev1alpha1.Pipeline)
	pipeline.ObjectMeta.Namespace = rig.Namespace
	g.Expect(ok).To(gomega.BeTrue())
	rig.K8s.Create(context.TODO(), pipeline)

	// Manually trigger the Pipeline by creating a PipelineInstance

	objs, err = e2e.LoadResourcesFromTestData("simple_build_pipeline_instance.yaml")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	instance, ok := objs[0].(*corev1alpha1.PipelineInstance)
	g.Expect(ok).To(gomega.BeTrue())
	t.Logf("instance pipeline: %s", instance.Spec.PipelineName)

	instance.ObjectMeta.Namespace = rig.Namespace

	err = rig.K8s.Create(context.TODO(), instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Eventually(func() error {
		i := &corev1alpha1.PipelineInstance{}
		return rig.K8s.Get(context.Background(), types.NamespacedName{Name: instance.Name, Namespace: rig.Namespace}, i)
	}, "30s").Should(gomega.Succeed())

	// Now the pipeline instance controller reconciles. It wants to associate the instance
	// with an Artifact matching the config provided. If there isn't already one matching the Source,
	// it creates one, without filling in the reference.

	g.Eventually(
		func() (*corev1alpha1.PipelineInstanceArtifact, error) {
			i := &corev1alpha1.PipelineInstance{}
			err := rig.K8s.Get(context.Background(), types.NamespacedName{Name: instance.Name, Namespace: rig.Namespace}, i)
			if err != nil {
				return nil, err
			}
			return i.Spec.Inputs["scm-upstream"].Artifact, err
		},
		"30s",
	).ShouldNot(gomega.BeNil())

	// Artifact controller triggers Git plugin to clone commit and put in local storage

	// Once artifact is "complete" we move on

	// First stage is triggered (PipelineStageInstance)
	// Spins up job with configured image and shell script
	// Build sidecar pulls local storage down into a shared volume
	// Waits for attached job to finish
	// Searches for outputs
	//   In this case we have a Docker image - it gets saved as a tarball and put in local storage
	//   Another artifact is created

}
