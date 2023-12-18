FROM golang:1.21.1

LABEL org.opencontainers.image.source='https://github.com/Jacobbrewer1/puppet-summary'
LABEL org.opencontainers.image.description="This is a puppet summary app written in Go that creates a summary of puppet runs."
LABEL org.opencontainers.image.licenses='GNU General Public License v3.0'

WORKDIR /puppet-summary

# Copy the binary from the build
COPY ./bin/app /puppet-summary/app

# Copy the assets over
COPY ./assets /puppet-summary/assets

RUN ["chmod", "+x", "./app"]

ENTRYPOINT ["/puppet-summary/app"]
