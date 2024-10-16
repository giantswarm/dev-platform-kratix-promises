FROM gsoci.azurecr.io/giantswarm/kratix-github-cli:0.1.0
USER root
RUN apk update && apk add --no-cache pre-commit=3.7.1-r0 curl=8.10.1-r0 npm=10.8.0-r0
RUN curl -fsSL https://github.com/gruntwork-io/boilerplate/releases/latest/download/boilerplate_linux_amd64 -o /usr/local/bin/boilerplate && chmod +x /usr/local/bin/boilerplate
RUN curl -fsSL https://github.com/norwoodj/helm-docs/releases/download/v1.14.2/helm-docs_1.14.2_Linux_x86_64.tar.gz -o /tmp/helm-docs.tar.gz && tar xzf /tmp/helm-docs.tar.gz && mv helm-docs /usr/local/bin/ && rm /tmp/helm-docs.tar.gz
USER nobody
COPY github-cli-template-values-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./_shared /_shared
COPY github-cli-template-values-pipeline/files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
