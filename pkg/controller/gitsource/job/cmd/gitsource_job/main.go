package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/milesbxf/puppeteer/pkg/apis"
	corev1alpha1 "github.com/milesbxf/puppeteer/pkg/apis/core/v1alpha1"
	git "gopkg.in/src-d/go-git.v4"
	gitconfig "gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	repository        = flag.String("repository-url", "", "URL of git repository to fetch/clone")
	branch            = flag.String("branch", "master", "Git branch to fetch/clone")
	storageAddress    = flag.String("storage-address", "http://localhost:9090", "URL of storage service")
	storagePathPrefix = flag.String("storage-prefix", "/v1alpha1/api/core/storage/upload", "API path to storage API")
	k8sNameRegexp     = regexp.MustCompile("[^a-zA-Z-.]")
	k8sclient         client.Client
)

func main() {
	flag.Parse()

	var err error
	k8sclient, err = setupK8sClient()
	if err != nil {
		log.Fatalf("Couldn't setup K8s client: %s", err.Error())
	}

	sha, err := getHeadCommit(*repository, *branch)
	if err != nil {
		log.Fatalf("Failed to fetch latest commits for branch %s on repo %s: %s", *branch, *repository, err.Error())
	}

	if artifactExists, err := artifactExistsForCommit(sha); err != nil {
		log.Fatal(err)
	} else if artifactExists {
		log.Printf("Artifact(s) already found for sha %s, not doing anything further", sha)
		return
	}

	var buf bytes.Buffer
	err = cloneRepoToTarReader(*repository, sha, &buf)
	if err != nil {
		log.Fatal(errors.New(fmt.Sprintf("Failed to clone repository and tar it: %s", err.Error())))
	}

	log.Printf("Compressed %d bytes of repository", buf.Len())

	artifactName := fmt.Sprintf("%s-%s", repoNametoK8sName(*repository), sha)

	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile("uploadfile", artifactName+".tar.gz")
	if err != nil {
		fmt.Println("error writing to buffer")
		log.Fatal(err)
	}
	_, err = io.Copy(fileWriter, &buf)
	if err != nil {
		log.Fatal(err)
	}

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	targetUrl := fmt.Sprintf("%s%s/%s", *storageAddress, *storagePathPrefix, artifactName)

	resp, err := http.Post(targetUrl, contentType, bodyBuf)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	resp_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Status)
	fmt.Println(string(resp_body))
	// If its a new commit
	//  - submit tarball to storage service
	//  - Create Artifact pointing at LocalStorageDirectory and commit

	// newArtifact := &corev1alpha1.Artifact{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      fmt.Sprintf("%s-%s", repoNametoK8sName(*repository), sha),
	// 		Namespace: "default",
	// 	},
	// 	Spec: corev1alpha1.ArtifactSpec{},
	// }
	// log.Printf("Creating artifact %s", newArtifact.Name)
	// err = k8sclient.Create(context.TODO(), newArtifact)
	// if err != nil {
	// 	log.Fatalf("Failed to look for existing artifacts: %s", err.Error())
	// }

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
func getHeadCommit(url, branch string) (string, error) {
	dir, err := ioutil.TempDir("", "gitpoll")
	if err != nil {
		return "", err
	}

	r, err := git.PlainInit(dir, true)
	r.CreateRemote(&gitconfig.RemoteConfig{
		Name:  "origin",
		URLs:  []string{*repository},
		Fetch: []gitconfig.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
	})
	if err != nil {
		return "", err
	}

	err = r.Fetch(&git.FetchOptions{})
	if err != nil {
		return "", err
	}
	remoteRef := plumbing.NewRemoteReferenceName("origin", branch)
	tip, err := r.Reference(plumbing.ReferenceName(remoteRef), true)
	if err != nil {
		return "", err
	}

	return tip.Hash().String(), nil
}

func artifactExistsForCommit(sha string) (bool, error) {
	selectorStr := fmt.Sprintf(
		"plugins.puppeteer.milesbryant.co.uk/v1alpha1/git/repository=%s,"+
			"plugins.puppeteer.milesbryant.co.uk/v1alpha1/git/commitsha=%s,",
		*repository, sha,
	)
	selector, _ := labels.Parse(selectorStr)
	artifacts := &corev1alpha1.ArtifactList{}
	err := k8sclient.List(context.TODO(), &client.ListOptions{
		LabelSelector: selector,
	}, artifacts)

	if err != nil {
		return false, errors.New(fmt.Sprintf("Failed to look for existing artifacts: %s", err.Error()))
	}
	return len(artifacts.Items) > 0, nil
}

func repoNametoK8sName(r string) string {
	name := strings.ToLower(r)
	name = strings.ReplaceAll(name, "https://", "")
	name = strings.ReplaceAll(name, "git@", "")
	name = k8sNameRegexp.ReplaceAllString(name, "-")
	return name
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

	log.Print("Compressing repository...")
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
