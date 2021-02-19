# Copyright © 2021 Mark Dumay. All rights reserved.
# Use of this source code is governed by The MIT License (MIT) that can be found in the LICENSE file.

#======================================================================================================================
# Build arguments
#======================================================================================================================
ARG BUILD_UID=1001
ARG BUILD_GID=1001
ARG BUILD_USER='restic'
ARG BUILD_FLAGS=''
ARG BUILD_VERSION
ARG BUILD_OS='linux'
ARG BUILD_ARCH='amd64'
ARG BUILD_TARGET='test'
ARG GOLANG_VERSION
ARG RESTIC_VERSION

#======================================================================================================================
# restic-unattended builder image
#======================================================================================================================
FROM golang:"${GOLANG_VERSION?version}" AS builder

# Compile restic-unattended binary from source for targeted OS and architectue
ARG BUILD_ARCH
ARG BUILD_OS
ARG BUILD_VERSION
COPY ./src /app/src/restic-unattended
WORKDIR /app/src/restic-unattended
RUN set -eu; \
    mkdir -p /app/bin; \
    go get -d -v ; \
    CGO_ENABLED=0 GOOS="${BUILD_OS}" GOARCH="${BUILD_ARCH}" go build \
        -ldflags="-w -s -X main.BuildVersion=${BUILD_VERSION}" -o /app/src/restic-unattended; \
    mv /app/src/restic-unattended/restic-unattended /app/bin/


#======================================================================================================================
# Final image with shell
#======================================================================================================================
ARG BUILD_RESTIC_VERSION
FROM restic/restic:"${RESTIC_VERSION}"

# Import helper script from build context and binaries from builder
COPY dbm/utils/harden_alpine.sh /usr/local/sbin/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/bin/restic-unattended /usr/local/bin/

# Create default non-root user and harden image
ARG BUILD_USER
ARG BUILD_UID
ARG BUILD_GID
ARG BUILD_FLAGS


RUN set -eu; \
    apk update -f; \
    apk --no-cache add -f shadow; \
    rm -rf /var/cache/apk/*; \
    chmod +x /usr/local/sbin/harden_alpine.sh; \
    /usr/local/sbin/harden_alpine.sh harden \
        -n "${BUILD_USER}" \
        -u "${BUILD_UID}" \
        -g "${BUILD_GID}" \
        "${BUILD_FLAGS}"; \
    rm -f /usr/local/sbin/harden_alpine.sh

# Run the container as non-root user
ARG BUILD_USER
USER "${BUILD_USER}":"${BUILD_USER}"

# Override entrypoint and start restic-unattended
# Note: use [""] syntax to avoid invoking potentially missing '/bin/sh'
ENTRYPOINT [""]
CMD ["restic-unattended"]