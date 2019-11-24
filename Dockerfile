FROM golang:alpine as build

WORKDIR /root

COPY . .

RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o main

from scratch

COPY --from=build /root/main .

ENTRYPOINT ["/main"]