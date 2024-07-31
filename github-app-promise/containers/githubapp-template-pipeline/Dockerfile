FROM gsoci.azurecr.io/giantswarm/kratix-base-cli
COPY execute-pipeline /bin/execute-pipeline
COPY ./shared /shared
COPY githubapp-template-pipeline/files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
