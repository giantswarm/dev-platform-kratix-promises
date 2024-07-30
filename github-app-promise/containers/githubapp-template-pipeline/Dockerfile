FROM giantswarm/kratix-github-cli
COPY execute-pipeline /bin/execute-pipeline
COPY --from=shared . /shared
CMD [ "bash", "-c", "/bin/execute-pipeline"]
ENTRYPOINT [] 
