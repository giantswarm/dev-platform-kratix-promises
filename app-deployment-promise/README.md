# App Deployment Promise

This promise allows to deploy a project created with [GitHub Template Repository Promise](../github-template-repo-promise/README.md) to an existing Kubernetes cluster.

## Usage

This resource is shaped after `HelmRelease` from Flux, as it is translated to one. Some fields are set automatically based on the status config map coming from the GitHub Template Repository Promise.

### Note about permissions

By default, kratix creates a new ServiceAccount for each namespace it runs resource request piplines in. Currently, this ServiceAccount has to be manually associated with needed roles.
