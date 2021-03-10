FROM golang:1.14.15-alpine3.13

# Install bash
RUN apk add --no-cache bash

# Set the Current Working Directory inside the container
WORKDIR /go/src/app

# Copy sources.
COPY . .

# Download all the dependencies.
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

ENV MAUTIC_DB_HOST="" \
    MAUTIC_DB_PORT="3306" \
    MAUTIC_DB_USER="" \
    MAUTIC_DB_PASSWORD="" \
    MAUTIC_DB_NAME="" \
    MAUTIC_TABLE_PREFIX=""

EXPOSE 9851

ADD /docker-entrypoint.sh /docker-entrypoint.sh

RUN set -x \
  && chmod +x /docker-entrypoint.sh

ENTRYPOINT ["/docker-entrypoint.sh"]

CMD ["mautic_exporter"]