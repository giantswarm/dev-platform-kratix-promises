FROM gsoci.azurecr.io/giantswarm/kratix-github-cli:0.1.1
COPY github-cli-clone-template-repo-pipeline/execute-pipeline /bin/execute-pipeline
COPY ./_shared /_shared
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
