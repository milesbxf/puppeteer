package gitartifactresolution

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	pluginsv1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/plugins/v1alpha1"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var rlog = logf.Log.WithName("resolver")

const (
	storageAddress    = "http://localhost:9090"
	storagePathPrefix = "/v1alpha1/api/core/storage"
)

func Resolve(res *pluginsv1alpha1.GitArtifactResolution) (*corev1alpha1.StorageReference, error) {
	logParams := []interface{}{"storage_id", res.Name}
	targetStatusUrl := fmt.Sprintf("%s%s/%s/status", storageAddress, storagePathPrefix, res.Name)
	resp, err := http.Get(targetStatusUrl)
	if err != nil {
		rlog.Error(err, "GET storage service", logParams...)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		var storageRef corev1alpha1.StorageReference
		d := json.NewDecoder(resp.Body)
		err := d.Decode(&storageRef)
		if err != nil {
			rlog.Error(err, "decoding response from storage")
			return nil, err
		}

		rlog.Info("got response from storage", "storage_response", storageRef)

		return &storageRef, nil
		// return decodeStorageRef(resp.Body)
	} else if resp.StatusCode != 404 {
		err := fmt.Errorf("unexpected status code from storage service: %d", resp.StatusCode)
		rlog.Error(err, "GET storage service", logParams...)
		return nil, err
	}

	rlog.Info("no storage found, cloning & creating", logParams...)

	var buf bytes.Buffer
	err = cloneRepoToTarReader(res.Spec.RepositoryURL, res.Spec.CommitSHA, &buf)
	if err != nil {
		rlog.Error(err, "Failed to clone repository and tar it", logParams...)
		return nil, err
	}

	rlog.Info("Compressed repository", append(logParams, "compressed_bytes", buf.Len())...)

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", res.Name+".tar.gz")
	if err != nil {
		rlog.Error(err, "writing to buffer", logParams...)
		return nil, err
	}
	_, err = io.Copy(fileWriter, &buf)
	if err != nil {
		rlog.Error(err, "writing to file", logParams...)
		return nil, err
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	targetUrl := fmt.Sprintf("%s%s/%s", storageAddress, storagePathPrefix, res.Name)
	resp, err = http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		rlog.Error(err, "POSTing to storage service", logParams...)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		bodybytes, _ := ioutil.ReadAll(resp.Body)
		err = fmt.Errorf("Got unexpected HTTP status code %d. Body: %s", resp.StatusCode, string(bodybytes))
		rlog.Error(err, "POSTing to storage service", logParams...)
		return nil, err
	}

	return decodeStorageRef(resp.Body)
}

func decodeStorageRef(body io.Reader) (*corev1alpha1.StorageReference, error) {
	var storageRef corev1alpha1.StorageReference
	d := json.NewDecoder(body)
	err := d.Decode(&storageRef)
	if err != nil {
		rlog.Error(err, "decoding response from storage")
		return nil, err
	}

	rlog.Info("got response from storage", "storage_response", storageRef)

	return &storageRef, nil
}

func cloneRepoToTarReader(repository, sha string, writer io.Writer) error {
	dir, err := ioutil.TempDir("", "gitclone")
	if err != nil {
		return err
	}
	r, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL:        repository,
		NoCheckout: true,
	})
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(sha),
	})
	if err != nil {
		return err
	}

	gzw := gzip.NewWriter(writer)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	rlog.Info("Compressing repository...")
	return filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
		// return on any error
		if err != nil {
			return err
		}

		// Don't try and tar directories
		if fi.Mode().IsDir() {
			return nil
		}

		newname := strings.TrimPrefix(strings.Replace(file, dir, "", -1), string(filepath.Separator))
		if len(newname) == 0 {
			return nil
		}

		// open file
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = newname
		header.Size = fi.Size()

		if fi.Mode()&os.ModeSymlink != 0 {
			header.Typeflag = tar.TypeSymlink

			target, err := os.Readlink(file)
			if err != nil {
				return err
			}
			header.Linkname = target
			return nil
		}

		rlog.V(8).Info("Compressing file", "filename", file, "size", header.Size, "type", fi.Mode().String())
		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		return nil
	})
}
