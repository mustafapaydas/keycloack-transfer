FROM golang:alpine AS builder
RUN apk update && apk add --no-cache git

COPY . /source

WORKDIR /source
WORKDIR /source


RUN go mod tidy
RUN go mod vendor

RUN go build -o /source/keycloack-transfer
FROM scratch



#COPY config.yaml /bin/keycloack-transfer/

COPY --from=builder /source/keycloack-transfer /bin/keycloack-transfer

ENTRYPOINT ["/bin/keycloack-transfer"]