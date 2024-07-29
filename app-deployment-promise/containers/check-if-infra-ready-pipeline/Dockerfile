FROM gsoci.azurecr.io/giantswarm/kratix-kubectl-cli:0.1.0
COPY check-if-infra-ready-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./shared /shared
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
