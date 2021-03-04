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
ARG BUILD_TARGET='test'
ARG ALPINE_VERSION
ARG GOLANG_VERSION
ARG RESTIC_VERSION

#======================================================================================================================
# restic builder image
#======================================================================================================================
FROM golang:"${GOLANG_VERSION?version}" AS builder1

# Compile restic binary from source for targeted OS and architectue
# Note: TARGETOS and TARGETARCH are only set automatically when using BuildKit
ARG TARGETOS
ARG TARGETARCH
ARG RESTIC_VERSION
ENV TARGETOS "${TARGETOS}"
ENV TARGETARCH "${TARGETARCH}"
WORKDIR /app/src/restic
RUN set -eu; \
    curl -sSL "https://github.com/restic/restic/archive/v${RESTIC_VERSION}.tar.gz" \
        | tar -C /app/src/restic/ -xz --strip 1; \
    [ -z "${TARGETOS}" ] && export TARGETOS=$(go env GOOS); \
    [ -z "${TARGETARCH}" ] && export TARGETARCH=$(go env GOARCH); \
    mkdir -p /app/bin; \
    go get -d -v ; \
    go run build.go --goos "${TARGETOS}" --goarch "${TARGETARCH}"; \
    mv /app/src/restic /app/bin/


#======================================================================================================================
# restic-unattended builder image
#======================================================================================================================
FROM golang:"${GOLANG_VERSION?version}" AS builder2

# Compile restic-unattended binary from source for targeted OS and architectue
# Note: TARGETOS and TARGETARCH are only set automatically when using BuildKit
ARG TARGETOS
ARG TARGETARCH
ENV TARGETOS "${TARGETOS}"
ENV TARGETARCH "${TARGETARCH}"
ARG BUILD_VERSION
COPY ./src /app/src/restic-unattended
WORKDIR /app/src/restic-unattended
RUN set -eu; \
    [ -z "${TARGETOS}" ] && export TARGETOS=$(go env GOOS); \
    [ -z "${TARGETARCH}" ] && export TARGETARCH=$(go env GOARCH); \
    mkdir -p /app/bin; \
    go get -d -v ; \
    CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" go build \
        -ldflags="-w -s -X main.BuildVersion=${BUILD_VERSION}" -o /app/src/restic-unattended; \
    mv /app/src/restic-unattended/restic-unattended /app/bin/


#======================================================================================================================
# Final image with shell
#======================================================================================================================
# ARG BUILD_RESTIC_VERSION
ARG ALPINE_VERSION
FROM alpine:"${ALPINE_VERSION?version}"

# Import helper script from build context and binaries from builder
COPY dbm/utils/harden_alpine.sh /usr/local/sbin/
COPY --from=builder1 /app/bin/restic /usr/local/bin/
COPY --from=builder2 /app/bin/restic-unattended /usr/local/bin/
COPY --from=builder2 /etc/passwd /etc/passwd
COPY --from=builder2 /etc/group /etc/group

# Create default non-root user and harden image
ARG BUILD_USER
ARG BUILD_UID
ARG BUILD_GID
ARG BUILD_FLAGS
RUN set -eu; \
    apk update -f; \
    apk --no-cache add -f shadow; \
    rm -rf /var/cache/apk/* /tmp; \
    chmod +x /usr/local/sbin/harden_alpine.sh; \
    /usr/local/sbin/harden_alpine.sh harden \
        -n "${BUILD_USER}" \
        -u "${BUILD_UID}" \
        -g "${BUILD_GID}" \
        -v /data/backup \
        -v /data/restore \
        -d /home/"${BUILD_USER}" \
        -d /tmp \
        "${BUILD_FLAGS}"; \
    rm -f /usr/local/sbin/harden_alpine.sh

# Run the container as non-root user
ARG BUILD_USER
USER "${BUILD_USER}":"${BUILD_USER}"

# Expose the backup and restore folders as volumes
VOLUME [ "/data/backup", "/data/restore" ]

# Define the healthcheck (production only)
ARG BUILD_TARGET
HEALTHCHECK --interval=5m --timeout=30s --retries=3 \
    CMD if [[ "${BUILD_TARGET}" == 'production' ]]; then restic-unattended check; else exit 0; fi

# Override entrypoint and start restic-unattended
# Note: use [""] syntax to avoid invoking potentially missing '/bin/sh'
ENTRYPOINT [""]
CMD ["restic-unattended"]