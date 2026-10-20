FROM fork-3x-ui:test
RUN apk add --no-cache nginx openssh-server openssh-sftp-server python3
COPY scripts/docker-test/*.sh /test/
COPY scripts/docker-test/start-panel.py x-ui.service.debian /test/
