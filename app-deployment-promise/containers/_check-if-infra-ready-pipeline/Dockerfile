FROM gsoci.azurecr.io/giantswarm/kratix-kubectl-cli:0.1.0
COPY check-if-infra-ready-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./_shared /_shared
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
