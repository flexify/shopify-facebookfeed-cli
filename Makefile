appname := shopify-facebookfeed
sources := main.go
build_dir := releases

build = GOOS=$(1) GOARCH=$(2) go build -o releases/$(appname)$(3)
buildWithName = GOOS=$(1) GOARCH=$(2) go build -ldflags="-w -s" -o releases/$(appname)__$(1)-$(2)$(3)
tar = cd $(build_dir) && tar -cvzf $(appname)_$(1)_$(2).tar.gz $(appname)$(3) && rm $(appname)$(3)
zip = cd $(build_dir) && zip $(appname)_$(1)_$(2).zip $(appname)$(3) && rm $(appname)$(3)

.PHONY: all windows darwin linux clean

all: windows darwin linux

clean:
	rm -rf $(build_dir)/

##### LINUX BUILDS #####
#linux: build/linux_arm.tar.gz build/linux_arm64.tar.gz build/linux_386.tar.gz build/linux_amd64.tar.gz
linux: build/linux_arm build/linux_arm64 build/linux_386 build/linux_amd64

build/linux_386.tar.gz: $(sources)
	$(call build,linux,386,)
	$(call tar,linux,386)

build/linux_amd64.tar.gz: $(sources)
	$(call build,linux,amd64,)
	$(call tar,linux,amd64)

build/linux_arm.tar.gz: $(sources)
	$(call build,linux,arm,)
	$(call tar,linux,arm)

build/linux_arm64.tar.gz: $(sources)
	$(call build,linux,arm64,)
	$(call tar,linux,arm64)

##

build/linux_386: $(sources)
	$(call buildWithName,linux,386,)

build/linux_amd64: $(sources)
	$(call buildWithName,linux,amd64,)

build/linux_arm: $(sources)
	$(call buildWithName,linux,arm,)

build/linux_arm64: $(sources)
	$(call buildWithName,linux,arm64,)

##### DARWIN (MAC) BUILDS #####
#darwin: build/darwin_amd64.tar.gz build/darwin_386.tar.gz
darwin: build/darwin_amd64 build/darwin_386

build/darwin_amd64.tar.gz: $(sources)
	$(call build,darwin,amd64,)
	$(call tar,darwin,amd64)

build/darwin_386.tar.gz: $(sources)
	$(call build,darwin,386,)
	$(call tar,darwin,386)

build/darwin_amd64: $(sources)
	$(call buildWithName,darwin,amd64,)

build/darwin_386: $(sources)
	$(call buildWithName,darwin,386,)

##### WINDOWS BUILDS #####

#windows: build/windows_386.zip build/windows_amd64.zip
windows: build/windows_386 build/windows_amd64

build/windows_386.zip: $(sources)
	$(call build,windows,386,.exe)
	$(call zip,windows,386,.exe)

build/windows_amd64.zip: $(sources)
	$(call build,windows,amd64,.exe)
	$(call zip,windows,amd64,.exe)


build/windows_386: $(sources)
	$(call buildWithName,windows,386,.exe)

build/windows_amd64: $(sources)
	$(call buildWithName,windows,amd64,.exe)
