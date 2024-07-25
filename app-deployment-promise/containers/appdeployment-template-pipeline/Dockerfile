FROM gsoci.azurecr.io/kratix-kubectl-cli:0.1.0
COPY execute-pipeline /bin/execute-pipeline
COPY --from=shared . /shared
COPY ./files/ /files/
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
