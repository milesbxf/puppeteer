package v1alpha1

import (
	"context"
	"testing"
	"time"

	"github.com/milesbxf/puppeteer/pkg/apis"
	pluginsv1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/plugins/v1alpha1"
	"github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/kubernetes/pkg/util/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var gitsourceName = types.NamespacedName{Name: "milesbxf-readme", Namespace: "default"}

const timeout = time.Second * 5

func TestGitSourceCreatesArtifact(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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
				URL:    "https://github.com/milesbxf/README",
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
		return
	}
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(context.TODO(), gitsource)

	found := &pluginsv1alpha1.GitSource{}

	g.Eventually(func() error { return c.Get(context.TODO(), gitsourceName, found) }, timeout).
		Should(gomega.Succeed())

}
