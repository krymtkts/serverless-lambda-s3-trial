export GO111MODULE=on
export GOOS=linux

.PHONY: build deploy gomodgen

ifdef profile
  AWSPROFILE=${profile}
else
  AWSPROFILE=default
endif

build:
	go mod tidy
	go build -ldflags="-s -w" -o bin/entry entry/main.go
	go build -ldflags="-s -w" -o bin/read-write ./read-write

deploy: build
	serverless deploy --verbose --aws-profile ${AWSPROFILE}

dryrun: build
	serverless deploy --verbose --noDeploy --aws-profile ${AWSPROFILE}
