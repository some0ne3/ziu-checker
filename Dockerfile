ARG GO_VERSION=1.21
FROM golang:${GO_VERSION}

WORKDIR /src

COPY main.go go.mod go.sum ./

RUN go build -o /bin/app .

CMD [ "/bin/app" ]
