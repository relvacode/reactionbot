FROM golang:latest as compiler
COPY . .
RUN go build -o /bin/reactionbot

FROM busybox

WORKDIR /
RUN mkdir /user-images

COPY --from=compiler /bin/reactionbot /bin/reactionbot

ENTRYPOINT ["/bin/reactionbot"]
