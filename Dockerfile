FROM golang:onbuild
COPY ./go-docker /usr/src/app
EXPOSE 9090