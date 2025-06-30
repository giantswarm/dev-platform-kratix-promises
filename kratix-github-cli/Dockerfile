FROM gsoci.azurecr.io/giantswarm/kratix-kubectl-cli:0.1.0
USER root
RUN apk update && apk add --no-cache github-cli=2.47.0-r4
RUN wget https://github.com/Link-/gh-token/releases/download/v2.0.5/linux-amd64 -O /usr/bin/gh-token && chmod +x /usr/bin/gh-token
USER nobody
