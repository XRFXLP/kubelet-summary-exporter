FROM golang:1.18 as build

WORKDIR /go/src/app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 make install

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/kubelet-summary-exporter /
CMD ["/kubelet-summary-exporter"]