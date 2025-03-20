FROM gsoci.azurecr.io/giantswarm/kratix-base-cli:0.1.1
COPY githubapp-template-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./_shared /_shared
COPY githubapp-template-pipeline/files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
