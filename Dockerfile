FROM ubuntu:latest
LABEL authors="piorus"

ENTRYPOINT ["top", "-b"]