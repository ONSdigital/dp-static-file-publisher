FROM golang:1.23.4-bullseye as build

WORKDIR /service
ADD . /service
CMD tail -f /dev/null

# TEST
FROM build as test

