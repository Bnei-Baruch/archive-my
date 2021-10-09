ARG work_dir=/go/src/github.com/Bnei-Baruch/archive-my
ARG build_number=dev
ARG mydb_url="postgres://user:password@host.docker.internal/mydb?sslmode=disable"
ARG mdb_url="postgres://user:password@host.docker.internal:5433/mdb?sslmode=disable"

FROM golang:1.16-alpine3.14 as build

LABEL maintainer="edoshor@gmail.com"

ARG work_dir
ARG build_number
ARG mydb_url
ARG mdb_url

ENV GOOS=linux \
	CGO_ENABLED=0 \
	MYDB_URL=${mydb_url} \
	MDB_URL=${mdb_url} \
	GIN_MODE=test

RUN apk update && \
    apk add --no-cache \
    git

WORKDIR ${work_dir}
COPY . .

RUN go test -v $(go list ./... | grep -v /models) \
    && go build -ldflags "-w -X github.com/Bnei-Baruch/archive-my/version.PreRelease=${build_number}"

FROM alpine:3.14

ARG work_dir
WORKDIR /app
COPY misc/wait-for /wait-for
COPY --from=build ${work_dir}/archive-my .

#COPY --from=build ${work_dir}/databases/mdb/migrations migrations/mdb

EXPOSE 8080

CMD ["./archive-my", "server"]
