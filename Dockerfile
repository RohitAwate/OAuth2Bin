FROM golang:1.11-alpine3.9
LABEL maintainer="Rohit Awate (https://github.com/RohitAwate)"

WORKDIR $GOPATH/src/github.com/RohitAwate/OAuth2Bin
COPY . .

RUN go build -o oauth2bin .
CMD [ "./oauth2bin" ]
