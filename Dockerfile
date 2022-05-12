FROM scratch

ADD echo-server /
COPY GeoLite2 /data/geolite2
ENV GEO_LITE_2_PATH=/data/geolite2/

CMD ["./echo-server"]