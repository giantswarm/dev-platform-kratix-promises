FROM gsoci.azurecr.io/giantswarm/kratix-base-cli:0.1.1
COPY provision-infrastructure-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./_shared /_shared
COPY provision-infrastructure-pipeline/files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
