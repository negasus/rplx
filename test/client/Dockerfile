FROM golang:1.12

WORKDIR /srv/client

ADD . .

RUN go build -o /client ./test/client

CMD ["/client"]