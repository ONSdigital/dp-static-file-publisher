FROM golang:1.18-stretch as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

# TEST
FROM build as test

