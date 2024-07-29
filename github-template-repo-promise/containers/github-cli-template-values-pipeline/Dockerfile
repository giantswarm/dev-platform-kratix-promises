FROM gsoci.azurecr.io/giantswarm/kratix-github-cli:0.1.0
COPY github-cli-template-values-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./shared /shared
COPY github-cli-template-values-pipeline/files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
