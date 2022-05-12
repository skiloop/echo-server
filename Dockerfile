FROM scratch

ADD echo-server /
RUN mkdir -p /data/geolite2
ENV GEO_LITE_2_PATH=/data/geolite2/

CMD ["./echo-server"]