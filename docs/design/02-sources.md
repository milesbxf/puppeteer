# Sources

A Source is a thing that kicks off a pipeline. The first one we'll implement is a git source.


## GitSource

```
apiVersion: plugins.puppeteer.milesbryant.co.uk/v1alpha1
kind: GitSource
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
  name: kubernetes-master
spec:
  repository:
    # remote URL to fetch/clone from
    url: https://github.com/kubernetes/kubernetes/
    branch: master
  clone:
    shallow: true
  poll:
    interval: 1m
```

This will set up a new job to poll kubernetes/kubernetes master branch for new commits every minute.
If there is a new commit, it will clone to storage (let's say local storage for now) and create a new artifact:


```
apiVersion: core.puppeteer.milesbryant.co.uk/v1alpha1
kind: Artifact
metadata:
  labels:
    controller-tools.k8s.io: "1.0"
    core.puppeteer.milesbryant.co.uk/artifact-id: kubernetes-master@1377108c084545e2279bd67eed56e54283572302
  name: git-kubernetes-master-1377108c084545e2279bd67eed56e54283572302
Spec:
  id: kubernetes-master@1377108c084545e2279bd67eed56e54283572302
  source: 
    apiGroupVersion: plugins.puppeteer.milesbryant.co.uk/v1alpha1.GitSource
    name: kubernetes-master
```

