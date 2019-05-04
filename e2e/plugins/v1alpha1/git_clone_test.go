package v1alpha1

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/milesbxf/puppeteer/pkg/apis"
	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	pluginsv1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/plugins/v1alpha1"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	gomegatypes "github.com/onsi/gomega/types"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/lithammer/shortuuid"
	"k8s.io/kubernetes/pkg/util/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var k8sNameRegexp = regexp.MustCompile("[^a-zA-Z-.]")

const timeout = time.Second * 5

func TestGitSourceCreatesArtifact(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	id := strings.ToLower(shortuuid.New())
	gitsourceName := types.NamespacedName{Name: "gitfixture-basic-" + id, Namespace: "default"}

	cfg, err := config.GetConfig()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	err = apis.AddToScheme(scheme.Scheme)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	c, err := client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
		Mapper: mapper,
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	gitsource := &pluginsv1alpha1.GitSource{
		ObjectMeta: metav1.ObjectMeta{Name: gitsourceName.Name, Namespace: gitsourceName.Namespace},
		Spec: pluginsv1alpha1.GitSourceSpec{
			Repository: pluginsv1alpha1.GitRepositoryOptions{
				URL:    "https://github.com/git-fixtures/basic",
				Branch: "master",
			},
			Poll: pluginsv1alpha1.PollOptions{
				IntervalMinutes: pointer.Int32Ptr(1),
			},
		},
	}

	err = c.Create(context.TODO(), gitsource)
	if apierrors.IsInvalid(err) {
		t.Logf("failed to create object, got an invalid object error: %v", err)
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), gitsource)

	found := &pluginsv1alpha1.GitSource{}

	g.Eventually(func() error { return c.Get(context.TODO(), gitsourceName, found) }, timeout).
		Should(gomega.Succeed())

	artifacts := &corev1alpha1.ArtifactList{}

	time.Sleep(5 * time.Minute)

	g.Eventually(func() error { return c.List(context.TODO(), &client.ListOptions{}, artifacts) }, 2*time.Minute).
		ShouldNot(BeEmptyArtifactList())

}

func repoNametoK8sName(r string) string {
	name := strings.ToLower(r)
	name = strings.Replace(name, "https://", "", -1)
	name = strings.Replace(name, "git@", "", -1)
	name = k8sNameRegexp.ReplaceAllString(name, "-")
	return name
}

type emptyArtifactListMatcher struct{}

func (m emptyArtifactListMatcher) Match(actual interface{}) (success bool, err error) {
	al, ok := actual.(*corev1alpha1.ArtifactList)
	if !ok {
		return false, fmt.Errorf("BeEmptyArtifactList matcher expects type ArtifactList, not %T", actual)
	}
	return gomega.BeEmpty().Match(al.Items)
}

func (m emptyArtifactListMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to be empty")
}

func (m emptyArtifactListMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to be empty")
}

func BeEmptyArtifactList() gomegatypes.GomegaMatcher {
	return emptyArtifactListMatcher{}
}
