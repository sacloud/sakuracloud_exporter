FROM golang:1.12 AS builder
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'

RUN  apt-get update && apt-get -y install \
        bash \
        git  \
        make \
      && apt-get clean \
      && rm -rf /var/cache/apt/archives/* /var/lib/apt/lists/*

ADD . /go/src/github.com/sacloud/sakuracloud_exporter
WORKDIR /go/src/github.com/sacloud/sakuracloud_exporter
RUN ["make", "build"]

#----------

FROM alpine:3.7
MAINTAINER Kazumichi Yamamoto <yamamoto.febc@gmail.com>
LABEL MAINTAINER 'Kazumichi Yamamoto <yamamoto.febc@gmail.com>'
RUN apk add --update ca-certificates

COPY --from=builder /go/src/github.com/sacloud/sakuracloud_exporter/bin/sakuracloud_exporter /usr/bin/

EXPOSE 9542

ENTRYPOINT ["/usr/bin/sakuracloud_exporter"]
