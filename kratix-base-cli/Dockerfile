FROM alpine:3.20.1
RUN apk update && apk add --no-cache yq-go=4.44.1-r2 git=2.45.2-r0 bash=5.2.26-r0 base64=0.5.2-r0 sed=4.9-r2 jq=1.7.1-r0 findutils=4.9.0-r5
RUN mkdir -p /kratix/input; mkdir /kratix/output; mkdir /kratix/metadata; chown -R nobody /kratix
RUN mkdir /tmp/; chown -R nobody /tmp
ENV HOME=/tmp
USER nobody
