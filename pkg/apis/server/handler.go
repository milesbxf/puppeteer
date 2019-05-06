package apis

import (
	"os"
	"path/filepath"

	"github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1/storage"
	"github.com/monzo/typhon"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("puppeteer-apiserver")

func Listen(addr string) error {

	// dir, err := ioutil.TempDir("", "puppeteer-storage")
	// if err != nil {
	// 	return err
	// }
	dir := filepath.Join(os.Getenv("HOME"), ".puppeteer/temp-storage")

	storage.Init(dir)

	_, err := typhon.Listen(storage.Router.Serve().
		Filter(typhon.ErrorFilter).
		Filter(typhon.H2cFilter), addr)
	return err
}
