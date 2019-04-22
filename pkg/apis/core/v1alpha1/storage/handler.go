package storage

import (
	"github.com/monzo/typhon"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var Router = typhon.Router{}

func init() {
	Router.POST("/v1alpha1/api/core/storage/upload/:id", handlePOSTStorageUpload)
}

func handlePOSTStorageUpload(req typhon.Request) typhon.Response {

	params := Router.Params(req)
	id := params["id"]

	log := logf.Log.WithName("handlePOSTStorageUpload")
	log.Info("Received file upload request", "id", id)

	req.ParseMultipartForm(32 << 30)
	uploadReader, handler, err := req.FormFile("uploadfile")
	if err != nil {
		log.Error(err, "Error parsing form data")
		return typhon.Response{Error: err}
	}
	defer uploadReader.Close()

	err = Store(uploadReader, handler.Filename)
	if err != nil {
		log.Error(err, "Error writing file")
		return typhon.Response{Error: err}
	}

	// Create a local storage object

	return req.Response(struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}
