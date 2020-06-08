#
# github.com/danieldin95/openlan-go
#

SHELL := /bin/bash

.ONESHELL:
.PHONY: linux linux/rpm darwin darwin/zip windows windows/zip test

## version
LSB = $(shell lsb_release -i -s)$(shell lsb_release -r -s)
VER = $(shell cat VERSION)

## declare flags
MOD = github.com/danieldin95/openlan-go/libol
LDFLAGS += -X $(MOD).Commit=$(shell git rev-list -1 HEAD)
LDFLAGS += -X $(MOD).Date=$(shell date +%FT%T%z)
LDFLAGS += -X $(MOD).Version=$(VER)

## declare directory
SD = $(shell pwd)
BD = $(SD)/build
LD = openlan-$(LSB)-$(VER)
WD = openlan-Windows-$(VER)
XD = openlan-Darwin-$(VER)

## all platform
all: linux windows darwin

pkg: linux/rpm windows/zip darwin/zip

clean:
	rm -rvf ./build

## prepare environment
env:
	@mkdir -p $(BD)

## linux platform
linux: linux/point linux/vswitch linux/ctrl

linux/ctrl: env
	cd controller && make linux

linux/point: env
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point ./main/point_linux

linux/vswitch: env
	go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-vswitch ./main/vswitch.go

linux/rpm: env
	@./packaging/spec.sh
	rpmbuild -ba packaging/openlan-ctrl.spec
	rpmbuild -ba packaging/openlan-point.spec
	rpmbuild -ba packaging/openlan-vswitch.spec
	@cp -rf ~/rpmbuild/RPMS/x86_64/openlan-*.rpm $(BD)

## cross build for windows
windows: windows/point windows/vswitch

windows/point: env
	GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.exe ./main/point_windows

windows/vswitch: env
	GOOS=windows GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-vswitch.exe ./main/vswitch.go

windows/zip: env windows
	@pushd $(BD)
	@rm -rf $(WD) && mkdir -p $(WD)
	@rm -rf $(WD).zip

	@cp -rvf $(SD)/packaging/resource/point.json.example $(WD)/point.json
	@cp -rvf $(BD)/openlan-point.exe $(WD)
	@cp -rvf $(BD)/openlan-vswitch.exe $(WD)

	zip -r $(WD).zip $(WD) > /dev/null
	@popd

windows/syso:
	rsrc -manifest main/point_windows/main.manifest -ico main/point_windows/main.ico  -o main/point_windows/main.syso

## cross build for osx
osx: darwin

darwin: env
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-point.darwin ./main/point_darwin
	GOOS=darwin GOARCH=amd64 go build -mod=vendor -ldflags "$(LDFLAGS)" -o $(BD)/openlan-vswitch.darwin ./main/vswitch.go

darwin/zip: env darwin
	@pushd $(BD)
	@rm -rf $(XD) && mkdir -p $(XD)
	@rm -rf $(XD).zip

	@cp -rvf $(SD)/packaging/resource/point.json.example $(XD)/point.json
	@cp -rvf $(BD)/openlan-point.darwin $(XD)
	@cp -rvf $(BD)/openlan-vswitch.darwin $(XD)

	zip -r $(XD).zip $(XD) > /dev/null
	popd

## unit test
test:
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/point
	go test -v -mod=vendor -bench=. github.com/danieldin95/openlan-go/libol
