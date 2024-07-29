# GitHub Template Repository Promise

Creates a new GitHub repository from a template repository using the GitHub API.

## Usage

A sample resource request is available in <resource-request.yaml>.

Before any new repository can be created, some input variables related to OCI and GitHub must be set.

```yaml
apiVersion: promise.platform.giantswarm.io/v1beta1
kind: githubrepo
metadata:
  name: my-go-project-1
spec:
  repository: ...
  registryInfoConfigMapRef: # a name reference to a ConfigMap that has to already exist
    # see below
    name: github-oci-registry-info
  githubTokenSecretRef: # a reference to a Secret that has to already exist; see below
    # it has to contain a key "GITHUB_TOKEN" with a github token that
    # has permissions to create repositories and read the template
    # repository
    name: dev-platform-gh-access
```

### RegistryInfoConfigMap structure

This config map can be shared across multiple `githubrepos`. If used with GHCR (currently the only tested setup), you can leave all the values as shown below and only provide the necessary secret. The ConfigMap must contain the following keys:

```yaml
registry_domain: ghcr.io # the domain of the OCI registry where images are uploaded
registry_username: build_bot # username used to authenticate to the registry
registry_cicd_secret_ref: "${{ secrets.GITHUB_TOKEN }}" # the name of the secret configured in the CI/CD project that can be used to push to the registry
registry_pull_secret_name: ghcr-pull-secret # the name of a docker-registry type secret that can be used to pull the images built by the projects into deployment clusters
```

The `registry_pull_secret_name` Secret can be created with the following command:

```bash
kubectl create secret docker-registry regcred --docker-server=<your-registry-server> --docker-username=<your-name> --docker-password=<your-pword> --docker-email=<your-email>
```

### Note about permissions

By default, kratix creates a new ServiceAccount for each namespace it runs resource request piplines in. Currently, this ServiceAccount has to be manually associated with needed roles.
