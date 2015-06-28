FROM golang:1.4

RUN go get -d github.com/garyburd/redigo/redis && go get github.com/nindalf/linkto

ENV LINKTO_HOSTNAME=http://nindalf.com
ENV LINKTO_WORDFILES="adjectives.txt animals.txt"
ENV LINKTO_PASSWORD=goop

WORKDIR /go/src/github.com/nindalf/linkto
RUN go build
EXPOSE 9091
ENTRYPOINT linkto
