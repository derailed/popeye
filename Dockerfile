FROM scratch
ENTRYPOINT ["/popeye"]
COPY popeye /
COPY --from=alpine:3.18.3 tmp/ /tmp/
