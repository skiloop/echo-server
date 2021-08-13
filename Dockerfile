FROM scratch

ADD echo-server /
ADD GeoLite2 /opt/geolite2
ENV GEO_LITE_2_PATH=/opt/geolite2/

CMD ["./echo-server"]