package v1alpha1_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/milesbxf/puppeteer/e2e"
	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"github.com/onsi/gomega"
	gomegafmt "github.com/onsi/gomega/format"
)

func TestSimpleBuildPipeline(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rig, err := e2e.NewTestRig()
	defer rig.TearDown()
	g.Expect(err).NotTo(gomega.HaveOccurred(), "setting up test rig")
	t.Logf("Test rig set up with namespace %s", rig.Namespace)

	_, pipeline := SetUpPipeline(t, rig, "simple_pipeline")

	// Now the pipeline instance controller reconciles. It wants to associate the instance
	// with an Artifact matching the config provided. If there isn't already one matching the Source,
	// it creates one, without filling in the reference.

	artifactName := WaitForPipelineArtifactName(t, rig, pipeline)

	WaitForArtifact(t, rig, artifactName)
	WaitForArtifactStorageReference(t, rig, artifactName)

	WaitForStagePhase(t, rig, pipeline.Name+"-build-1", corev1alpha1.StageInProgress)
	WaitForTaskPhase(t, rig, pipeline.Name+"-build-1-build-image-1", corev1alpha1.TaskInProgress)

	WaitForTaskPhase(t, rig, pipeline.Name+"-build-1-build-image-1", corev1alpha1.TaskComplete)
	WaitForStagePhase(t, rig, pipeline.Name+"-build-1", corev1alpha1.StageComplete)
}

func TestBrokenPipeline(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rig, err := e2e.NewTestRig()
	defer rig.TearDown()
	g.Expect(err).NotTo(gomega.HaveOccurred(), "setting up test rig")
	t.Logf("Test rig set up with namespace %s", rig.Namespace)

	_, pipeline := SetUpPipeline(t, rig, "broken_pipeline")

	artifactName := WaitForPipelineArtifactName(t, rig, pipeline)

	WaitForArtifact(t, rig, artifactName)
	WaitForArtifactStorageReference(t, rig, artifactName)

	WaitForStagePhase(t, rig, pipeline.Name+"-build-1", corev1alpha1.StageInProgress)
	WaitForTaskPhase(t, rig, pipeline.Name+"-build-1-build-image-1", corev1alpha1.TaskInProgress)

	WaitForTaskPhase(t, rig, pipeline.Name+"-build-1-build-image-1", corev1alpha1.TaskError)
	WaitForStagePhase(t, rig, pipeline.Name+"-build-1", corev1alpha1.StageError)
}

func TestPipelineWithStorage(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	rig, err := e2e.NewTestRig()
	defer rig.TearDown()
	g.Expect(err).NotTo(gomega.HaveOccurred(), "setting up test rig")
	t.Logf("Test rig set up with namespace %s", rig.Namespace)

	_, pipeline := SetUpPipeline(t, rig, "storage_pipeline")

	artifactName := WaitForPipelineArtifactName(t, rig, pipeline)

	WaitForArtifact(t, rig, artifactName)
	WaitForArtifactStorageReference(t, rig, artifactName)

	WaitForStagePhase(t, rig, pipeline.Name+"-build-1", corev1alpha1.StageInProgress)
	WaitForTaskPhase(t, rig, pipeline.Name+"-build-1-build-image-1", corev1alpha1.TaskInProgress)

	WaitForTaskPhase(t, rig, pipeline.Name+"-build-1-build-image-1", corev1alpha1.TaskComplete)
	WaitForStagePhase(t, rig, pipeline.Name+"-build-1", corev1alpha1.StageComplete)

	gomegafmt.TruncatedDiff = false

	output := GetPodOutput(t, rig, pipeline.Name+"-build-1-build-image-1")
	if e2e.UpdateGolden() {
		err := ioutil.WriteFile("testdata/storage_pipeline.golden", []byte(output), 0644)
		g.Expect(err).NotTo(gomega.HaveOccurred(), "writing golden file")
	}
	goldenBytes, err := ioutil.ReadFile("testdata/storage_pipeline.golden")
	g.Expect(err).NotTo(gomega.HaveOccurred(), "reading golden file")
	golden := strings.TrimSpace(string(goldenBytes))

	g.Expect(output).To(gomega.Equal(golden))

}
