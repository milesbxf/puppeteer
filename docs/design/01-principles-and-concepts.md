# 👩‍⚖️ Principles

## ☸︎ Use of Kubernetes

The entire system should be modeled around the Kubernetes API, using custom resources, APIs and controllers. Kubernetes is widely used and helps solve a lot of difficult distributed systems problems. It includes a powerful job scheduler and access control mechanism, encourages building highly resilient applications, and has been shown to scale.

Using custom resource definitions (CRDs) mean that we can debug the state of the system just by using kubectl (or the Kubernetes API). It also means that end users can create really highly granular RBAC policies.

Writing controllers that operate on these CRDs keeps the system highly decoupled yet (if we design it right!) cohesive, meaning that each part does one thing well. This should make it easy to test parts of the system in isolation.

This will also make the system highly extensible. Almost everything outside of the core concepts should be written and deployed as a plugin - this concept works quite well in Caddy and CoreDNS.
Ideally, someone would be able to dynamically add or update a plugin just by deploying a Kubernetes application.


## Instrumentation

The system should be easy to instrument and analyse, both from an observability and monitoring perspective, and from a data/business analytics perspective. I've always admired the approach of the [Hygieia project](http://hygieia.github.io/Hygieia/screenshots.html) - effective engineering organisations like to objectively track the effectiveness of their delivery pipelines, identify bottlenecks, etc.


## Security, access control and provenance

It's important to control who can update pipeline configuration and trigger pipelines on a per-resource basis. Also, e.g. non-admins shouldn't be able to create intermediate resources.
As stated above, we can use RBAC to enforce this, perhaps coming with a sample set of roles to apply to users.

Many organisations have change control and audit requirements around their delivery systems. One key idea is to be able to track a particular artifact back to its source, and verify that it has passed a number of quality gates. See this [article about code provenance checking at Uber](https://medium.com/uber-security-privacy/code-provenance-application-security-77ebfa4b6bc5) for more


# 👩‍🏫 Concepts

## 🗿 Artifacts

Everything in Puppeteer should revolve around the concept of an Artifact. An Artifact is a immutable object passed between different processes in the system, which may produce other Artifacts as a result. Examples of some Artifacts:
* A git repository at a particular commit hash
* A Docker image (pinned to a specific SHA)
* A binary built on a pipeline

#### Lifecycle

An Artifact can (but not always) start off as a request for an Artifact. For instance, if we trigger a pipeline with a particular Git commit, we'll create an Artifact that references that commit as a source, but doesn't reference any storage. This is an "unresolved" Artifact.

The Artifact controller will reconcile unresolved Artifacts, and do what it needs to do to resolve them and fill out the reference - this is detailed in the next section, Sources.

Once an Artifact has a reference and it is "resolved", it is ready for use in build stages and other pipelines.

Later on, we may want to introduce some form of garbage collection where we clean up the underlying storage of old Artifacts, so that may introduce a new phase.


#### Properties

The main property that Artifacts have is that they're uniquely identifiable and, if they're "resolved", they're immutable. This means a Git repository pinned to a commit SHA (rather than a tag), and Docker images pinned to a SHA.
Artifacts will typically (but not always) have some kind of state associated with them. For example, a Git repository will have the filesystem tree of repository contents (and Git metadata), the Docker image will have the image layers.
It seems sensible to not use Kubernetes to store this state - most typical use cases would exceed the [limits of etcd](https://github.com/etcd-io/etcd/blob/master/Documentation/dev-guide/limit.md) quite quickly.

Instead state should be stored in some other kind of backend. It makes sense to implement these backends as plugins, to make it easy to write new backends. Some examples of storage backends:
* Local filesystem - storage service would just store and return files locally. Good for testing, not so scalable or available
* Object storage ala Minio/AWS S3 or Google Cloud Storage - these are cheap to store lots of data (and actually have pretty good latency when running in the cloud)
* Docker registries
