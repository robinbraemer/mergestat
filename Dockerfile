# syntax=docker/dockerfile:1-experimental

FROM golang:1.19-alpine3.16 AS builder
RUN set -x && apk add --no-cache cmake git make gcc libtool g++ openssl1.1-compat-dev
COPY scripts/install_libgit2.sh install_libgit2.sh
RUN ./install_libgit2.sh
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/usr/share/pkgconfig/libgit2/lib/pkgconfig/ make

FROM zricethezav/gitleaks:v8.15.2 AS gitleaks

FROM alpine:3.16
RUN set -x && apk add --no-cache curl postgresql-client ca-certificates git go

# copy over migrations
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar xvz
RUN mv migrate /usr/local/bin
COPY migrations migrations

COPY scripts/docker-init-entrypoint.sh docker-init-entrypoint.sh

# install trivy binary for `TRIVY_REPO_SCAN` sync type
RUN curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.31.3

# install syft binary
RUN curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin v0.58.0

# install the gosec binary
RUN curl -sfL https://raw.githubusercontent.com/securego/gosec/master/install.sh | sh -s -- -b /usr/local/bin/ v2.14.0

# install gitleaks binary from gitleaks image
COPY --from=gitleaks /usr/bin/gitleaks /usr/local/bin/gitleaks

# install python3 and yelp detect-secrets
ENV PYTHONUNBUFFERED=1
RUN apk add --update --no-cache python3 && ln -sf python3 /usr/bin/python
RUN python3 -m ensurepip
RUN pip3 install --no-cache --upgrade pip setuptools
RUN pip3 install detect-secrets

# for pprof and prom metrics over http
EXPOSE 8080

COPY --from=builder /src/.build/worker /worker
ENTRYPOINT [ "/worker" ]
