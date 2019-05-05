package e2e

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lithammer/shortuuid"
	"github.com/milesbxf/puppeteer/pkg/apis"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type TestRig struct {
	K8s       client.Client
	Namespace string
}

func NewTestRig() (*TestRig, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
	if err != nil {
		return nil, err
	}

	err = apis.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	c, err := client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
		Mapper: mapper,
	})
	if err != nil {
		return nil, err
	}

	id := strings.ToLower(shortuuid.New())
	testNamespace := "puppeteer-e2e-test-" + id

	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: testNamespace,
			Labels: map[string]string{
				"puppeteer-e2e-test": "true",
			},
		},
	}

	err = c.Create(context.TODO(), &ns)
	if err != nil {
		return nil, err
	}

	return &TestRig{
		K8s:       c,
		Namespace: testNamespace,
	}, nil
}

func (t *TestRig) TearDown() {
	ns := corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: t.Namespace}}
	t.K8s.Delete(context.Background(), &ns)
}

func LoadResourcesFromTestData(name string) ([]runtime.Object, error) {
	f, err := os.Open(filepath.Join("testdata", name))
	if err != nil {
		return nil, fmt.Errorf("open file %s: %s", name, err.Error())
	}
	objects := []runtime.Object{}

	reader := bufio.NewReader(f)
	// use K8s YAML reader to split up documents (1 doc should == 1 object)
	decoder := yaml.NewYAMLReader(reader)

	for {
		bytes, err := decoder.Read()

		if len(bytes) == 0 {
			// no more documents
			return objects, nil
		}

		if err != nil {
			return nil, fmt.Errorf("decode doc from %s: %s", name, err.Error())
		}

		obj, _, err := scheme.Codecs.UniversalDeserializer().Decode(bytes, nil, nil)

		if err != nil {
			return nil, fmt.Errorf("parse resource from bytes: %s", err.Error())
		}

		objects = append(objects, obj)
	}
}
