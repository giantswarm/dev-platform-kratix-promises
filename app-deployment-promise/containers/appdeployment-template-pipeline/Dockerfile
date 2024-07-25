FROM gsoci.azurecr.io/giantswarm/kratix-kubectl-cli:0.1.0
COPY appdeployment-template-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./shared /shared
COPY appdeployment-template-pipeline/files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
