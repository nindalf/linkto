FROM golang:1.4
MAINTAINER Krishna Sundarram <krishna.sundarram@gmail.com>

RUN go get -d github.com/garyburd/redigo/redis
ADD . /go/src/github.com/nindalf/linkto
WORKDIR /go/src/github.com/nindalf/linkto
RUN go install

EXPOSE 9091
ENTRYPOINT linkto
