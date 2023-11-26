FROM golang:alpine AS bootstrap

WORKDIR /build
COPY . .
RUN go build -o /lambda

# - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -

FROM alpine AS lambda

RUN apk add curl
WORKDIR /
COPY --from=bootstrap /lambda /lambda

HEALTHCHECK CMD curl --fail http://localhost:3000/health 2> /dev/null || exit 1 

ENTRYPOINT ["/lambda"]

