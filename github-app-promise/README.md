# GitHubApp Promise

This promise combines the [GitHubRepo](../github-template-repo-promise/README.md) and [AppDeployment](../app-deployment-promise/README.md) promises to provide a complete GitHub App deployment solution.

The purpose of it is:

- to demonstrate how one can combine smaller promises into bigger ones, that can automate part of the work required to create lower level promises separately,
- to make the bootstrapping of a new project easier, so that a new `GitHubRepo` and `AppDeployment` can be created in one go.

## Implementation

This promise renders the two other promises: `GitHubRepo` and `AppDeployment` and sets the `project-info` ConfigMap rendered by the `GitHubRepo` promise as the input for the `AppDeployment` promise.
