FROM golang:1.12

WORKDIR /srv/app

ADD . .

RUN go build -o /app ./test

CMD ["/app"]