FROM --platform=${BUILDPLATFORM} golang:alpine as compiler
ARG TARGETOS
ARG TARGETARCH
ENV CGO_ENABLED=0

WORKDIR /build

COPY . .

RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build github.com/relvacode/reactionbot

FROM --platform=${TARGETPLATFORM} alpine

WORKDIR /
RUN mkdir /user-images

COPY --from=compiler /build/reactionbot /bin/reactionbot

ENTRYPOINT ["/bin/reactionbot"]
