FROM golang

ARG app_env
ENV APP_ENV $app_env

COPY src /go/src
ADD fresh.conf /go/src
ADD dep.sh /go/src
ADD revision.txt /go/src
ADD version.txt /go/src

WORKDIR /go/src
COPY go.mod go.sum ./
RUN go mod tidy

RUN \
    sh ./dep.sh

RUN go build -o galcon .

CMD if [ ${APP_ENV} = production ]; \
	then \
	./galcon; \
	else \
	go get github.com/pilu/fresh && \
	fresh -c fresh.conf; \
	fi

EXPOSE 8080
