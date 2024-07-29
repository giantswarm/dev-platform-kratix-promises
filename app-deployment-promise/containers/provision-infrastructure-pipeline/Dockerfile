FROM gsoci.azurecr.io/giantswarm/kratix-base-cli:0.1.0
COPY provision-infrastructure-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./shared /shared
COPY provision-infrastructure-pipeline/files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
