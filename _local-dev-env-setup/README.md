# How to setup local development environment

1. Bootstrap a new `kind` cluster and deploy `kratix` and its dependencies:

```
./bootstrap-kratix-on-kind.sh
```

2. Download [`gitea`](https://about.gitea.com/products/gitea/)
3. Configure `gitea` to [use its built-in HTTPS support](https://docs.gitea.com/administration/https-setup).
4. Configure gitea: register a new user, create a repo for `gitops` and create an access token with write permissions. Save the token into `GITEA_TOKEN` env var.
5. Bootstrap flux:

```
flux bootstrap gitea --hostname https://172.18.0.1:3000 --owner [USER] --repository kratix --ca-file ~/tools/gitea/custom/cert.pem --path flux/kind-mc
```

6. Create kratix `gitstatestore` using the same repo and token as flux:

```yaml
apiVersion: platform.kratix.io/v1alpha1
kind: GitStateStore
metadata:
  name: gitea
spec:
  authMethod: basicAuth
  branch: main
  gitAuthor:
    name: kratix
  secretRef:
    name: flux-system
    namespace: flux-system
  url: https://172.18.0.1:3000/[USER]/kratix.git
```

7. Create kratix `destination` for the kind cluster itself

```yaml
apiVersion: platform.kratix.io/v1alpha1
kind: Destination
metadata:
  labels:
    environment: dev
  name: kind-mc
spec:
  filepath:
    mode: none
  stateStoreRef:
    kind: GitStateStore
    name: gitea
```
