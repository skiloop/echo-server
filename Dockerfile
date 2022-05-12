FROM scratch

ADD echo-server /app/echo-server
ENV GEO_LITE_2_PATH=/app/config/geolite2/
CMD ["/app/echo-server"]