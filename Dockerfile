FROM scratch

ADD echo-server /app
ENV GEO_LITE_2_PATH=/app/config/geolite2/
CMD ["/app/echo-server"]