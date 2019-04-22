package apis

import (
	"github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1/storage"
	"github.com/monzo/typhon"
)

func Listen(addr string) error {
	_, err := typhon.Listen(storage.Router.Serve().
		Filter(typhon.ErrorFilter).
		Filter(typhon.H2cFilter), addr)
	return err
}
