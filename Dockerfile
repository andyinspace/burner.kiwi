FROM ubuntu:20.04
ARG TARGETARCH
COPY burnerkiwi.${TARGETARCH} burnerkiwi
EXPOSE 8080 25 9091
ENTRYPOINT [ "/burnerkiwi" ]