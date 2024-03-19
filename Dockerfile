FROM golang:1.22.0-alpine3.19 AS builder

RUN apk upgrade

RUN echo "alpine:x:1000:0:alpine:/tmp:/sbin/nologin" >> /etc/passwd

ARG SwaggerUiDistVer=5.12.0
ARG GitCommitId

# download swagger ui
RUN wget -O /tmp/swagger-resources.tgz \
  https://registry.npmjs.org/swagger-ui-dist/-/swagger-ui-dist-$SwaggerUiDistVer.tgz \
  && cd /tmp && tar -xf swagger-resources.tgz && rm -rf swagger-resources.tgz \
  && mv package swagger-ui-resources

# Replace default swagger json with our api json file
RUN sed -i 's#https:\/\/petstore.swagger.io\/v2\/swagger.json#..\/swagger.json#g'  /tmp/swagger-ui-resources/swagger-initializer.js

WORKDIR /go/src/trustdoc

# Caching all external dependencies in docker context
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the source from the current directory
COPY . .

# Dependencies for project which use cgo like etherium
RUN apk update && apk add --no-progress --no-cache gcc musl-dev

RUN go build -ldflags "-X github.com/vposham/trustdoc/handler/base.GitCommitId=$GitCommitId" \
    -o /go/src/trustdoc/app -tags musl

FROM alpine:3.19.1
# Add tzdata for timezone information in logs
RUN apk add tzdata

WORKDIR /go/src/trustdoc

COPY --from=builder /tmp/swagger-ui-resources docs/swagger-ui-dist
COPY --from=builder /go/src/trustdoc/app .
COPY --from=builder /go/src/trustdoc/config ./config
COPY --from=builder /go/src/trustdoc/docs ./docs

EXPOSE 8080
ENTRYPOINT ["./app"]
