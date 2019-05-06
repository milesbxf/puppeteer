package v1alpha1_test

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/milesbxf/puppeteer/e2e"
	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SetUpPipeline(t *testing.T, rig *e2e.TestRig, name string) (*corev1alpha1.PipelineConfig, *corev1alpha1.Pipeline) {
	g := gomega.NewGomegaWithT(t)

	config_file := name + "_config.yaml"
	pipeline_file := name + ".yaml"

	// Load and create a simple Pipeline Config

	objs, err := e2e.LoadResourcesFromTestData(config_file)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "loading pipeline config yaml")

	pipelineConfig, ok := objs[0].(*corev1alpha1.PipelineConfig)
	pipelineConfig.ObjectMeta.Namespace = rig.Namespace
	g.Expect(ok).To(gomega.BeTrue(), "loading pipeline config obj")
	err = rig.K8s.Create(context.TODO(), pipelineConfig)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "creating pipeline config obj")
	t.Logf("Created pipeline object %s", pipelineConfig.Name)

	// Manually trigger a Pipeline

	objs, err = e2e.LoadResourcesFromTestData(pipeline_file)
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

	return pipelineConfig, pipeline
}

func GetArtifactFromPipeline(c client.Client, name, namespace string) (*corev1alpha1.PipelineArtifact, error) {
	p := &corev1alpha1.Pipeline{}
	err := c.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, p)
	if err != nil {
		return nil, err
	}
	return p.Spec.Inputs["scm-upstream"].Artifact, err
}

func WaitForPipelineArtifactName(t *testing.T, rig *e2e.TestRig, pipeline *corev1alpha1.Pipeline) string {
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() (*corev1alpha1.PipelineArtifact, error) {
		return GetArtifactFromPipeline(rig.K8s, pipeline.Name, rig.Namespace)
	}, "30s").
		ShouldNot(gomega.BeNil(), "waiting for pipeline artifact name")

	a, err := GetArtifactFromPipeline(rig.K8s, pipeline.Name, rig.Namespace)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "getting artifact from pipeline")
	t.Logf("Got name of artifact for pipeline input: '%s'", a.Name)

	return a.Name
}

func WaitForArtifact(t *testing.T, rig *e2e.TestRig, artifactName string) *corev1alpha1.Artifact {
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() error {
		artifact := &corev1alpha1.Artifact{}
		return rig.K8s.Get(context.Background(), types.NamespacedName{Name: artifactName, Namespace: rig.Namespace}, artifact)
	}, "30s").
		Should(gomega.Succeed(), "waiting for artifact")

	artifact := &corev1alpha1.Artifact{}
	err := rig.K8s.Get(context.Background(), types.NamespacedName{Name: artifactName, Namespace: rig.Namespace}, artifact)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "getting artifact")
	t.Logf("Got artifact for pipeline instance input: '%s'", artifact.Name)
	return artifact
}

func GetStorageReferenceFromArtifact(c client.Client, name, namespace string) (*corev1alpha1.StorageReference, error) {
	artifact := &corev1alpha1.Artifact{}
	err := c.Get(context.Background(), types.NamespacedName{Name: name, Namespace: namespace}, artifact)
	if err != nil {
		return nil, err
	}
	return artifact.Status.Reference, nil

}

func WaitForArtifactStorageReference(t *testing.T, rig *e2e.TestRig, artifactName string) *corev1alpha1.StorageReference {
	g := gomega.NewGomegaWithT(t)

	g.Eventually(func() (*corev1alpha1.StorageReference, error) {
		return GetStorageReferenceFromArtifact(rig.K8s, artifactName, rig.Namespace)
	}, "2m").
		ShouldNot(gomega.BeNil(), "getting artifact reference")

	sr, err := GetStorageReferenceFromArtifact(rig.K8s, artifactName, rig.Namespace)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "getting storage reference")
	return sr
}

func WaitForStagePhase(t *testing.T, rig *e2e.TestRig, stageName string, phase corev1alpha1.StagePhase) {
	g := gomega.NewGomegaWithT(t)

	g.Eventually(
		func() (corev1alpha1.StagePhase, error) {
			stage := &corev1alpha1.Stage{}
			err := rig.K8s.Get(context.Background(), types.NamespacedName{Name: stageName, Namespace: rig.Namespace}, stage)
			if err != nil {
				return "", err
			}
			return stage.Status.Phase, nil
		},
		"30s",
	).Should(gomega.Equal(phase), "waiting for pipeline stage "+string(phase))
	t.Logf("Pipeline stage is %s", phase)
}

func WaitForTaskPhase(t *testing.T, rig *e2e.TestRig, taskName string, phase corev1alpha1.TaskPhase) {
	g := gomega.NewGomegaWithT(t)

	g.Eventually(
		func() (corev1alpha1.TaskPhase, error) {
			task := &corev1alpha1.Task{}
			err := rig.K8s.Get(context.Background(), types.NamespacedName{Name: taskName, Namespace: rig.Namespace}, task)
			if err != nil {
				return "", err
			}
			return task.Status.Phase, nil
		},
		"30s",
	).Should(gomega.Equal(phase), "waiting for pipeline task "+string(phase))
	t.Logf("Pipeline task is %s", phase)
}

func GetPodOutput(t *testing.T, rig *e2e.TestRig, taskName string) string {
	g := gomega.NewGomegaWithT(t)
	podList, err := rig.ClientGoK8s.CoreV1().Pods(rig.Namespace).List(metav1.ListOptions{
		LabelSelector: "puppeteer.milesbryant.co.uk/task-name=" + taskName,
	})
	g.Expect(err).NotTo(gomega.HaveOccurred(), "looking up pod")
	g.Expect(podList.Items).To(gomega.HaveLen(1), "looking up pod")

	podName := podList.Items[0].Name

	req := rig.ClientGoK8s.CoreV1().RESTClient().Get().
		Namespace(rig.Namespace).
		Name(podName).
		Resource("pods").
		SubResource("log").
		Param("follow", "false")

	readCloser, err := req.Stream()
	g.Expect(err).NotTo(gomega.HaveOccurred(), "requesting pod logs")

	defer readCloser.Close()
	b, err := ioutil.ReadAll(readCloser)
	g.Expect(err).NotTo(gomega.HaveOccurred(), "reading pod logs")

	return string(b)
}
