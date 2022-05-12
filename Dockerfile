FROM scratch

ADD echo-server /
ENV GEO_LITE_2_PATH=/data/geolite2/
CMD ["./echo-server"]