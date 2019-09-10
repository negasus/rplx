FROM golang:1.12

WORKDIR /srv/server

ADD . .

RUN go build -o /server ./test/server

EXPOSE 2000
EXPOSE 3000

CMD ["/server"]