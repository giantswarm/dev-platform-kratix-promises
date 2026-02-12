FROM gsoci.azurecr.io/giantswarm/kratix-kubectl-cli:0.1.1
COPY appdeployment-template-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./_shared /_shared
COPY appdeployment-template-pipeline/files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
