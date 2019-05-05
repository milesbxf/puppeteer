package storage

import (
	"context"
	"fmt"

	"github.com/milesbxf/puppeteer/pkg/apis"
	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	"github.com/monzo/typhon"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var Router = typhon.Router{}
var log = logf.Log.WithName("storageHandler")

func init() {
	Router.POST("/v1alpha1/api/core/storage/upload/:id", handlePOSTStorageUpload)
}

// TODO: urgh, get rid of this package-level global crap and move into a struct
var rootPath string = ""

func Init(storageRootPath string) {
	rootPath = storageRootPath
	log.Info("Initialised storage directory", "storage_directory", storageRootPath)
}

func handlePOSTStorageUpload(req typhon.Request) typhon.Response {

	params := Router.Params(req)
	id := params["id"]

	log.Info("Received file upload request", "id", id)

	req.ParseMultipartForm(32 << 30)
	uploadReader, handler, err := req.FormFile("uploadfile")
	if err != nil {
		log.Error(err, "Error parsing form data")
		return typhon.Response{Error: err}
	}
	defer uploadReader.Close()

	err = Store(rootPath, uploadReader, handler.Filename)
	if err != nil {
		log.Error(err, "Error writing file")
		return typhon.Response{Error: err}
	}

	// Create a local storage object
	obj, err := createLocalStorageObj(id, handler.Filename)
	if err != nil {
		log.Error(err, "Error creating LocalStorage object")
		return typhon.Response{Error: err}
	}

	sgv := corev1alpha1.SchemeGroupVersion

	resp := req.Response(corev1alpha1.StorageReference{
		Status:               corev1alpha1.StorageStatusPresent,
		GroupVersionResource: fmt.Sprintf("localstorage.%s.%s", sgv.Version, sgv.Group),
		ID:                   obj.ObjectMeta.Name,
	})
	resp.StatusCode = 201
	return resp
}
func createLocalStorageObj(id, key string) (*corev1alpha1.LocalStorage, error) {
	k8sclient, err := setupK8sClient()
	if err != nil {
		return nil, err
	}
	obj := &corev1alpha1.LocalStorage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id,
			Namespace: "default",
		},
		Spec: corev1alpha1.LocalStorageSpec{
			Key: key,
		},
	}
	err = k8sclient.Create(context.TODO(), obj)
	return obj, err
}

func setupK8sClient() (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}

	err = apis.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, err
	}

	mapper, err := apiutil.NewDiscoveryRESTMapper(cfg)
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
	return c, nil
}
