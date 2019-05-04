package v1alpha1_test

import (
	"context"
	"testing"

	"github.com/milesbxf/puppeteer/e2e"
	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"github.com/onsi/gomega"
)

func TestSimpleBuildPipeline(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rig, err := e2e.NewTestRig()
	defer rig.TearDown()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Load and create pipeline resource
	objs, err := e2e.LoadResourcesFromTestData("simple_build_pipeline.yaml")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	pipeline, ok := objs[0].(*corev1alpha1.Pipeline)
	g.Expect(ok).To(gomega.BeTrue())
	rig.K8s.Create(context.TODO(), pipeline)

	// Create a PipelineInstance object
	// Artifact created referencing git SHA but no storage reference

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
