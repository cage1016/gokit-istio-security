# Authzopa mixer custom adpater

## Content
- [Authzopa mixer custom adpater](#authzopa-mixer-custom-adpater)
  - [Content](#content)
    - [Before Start](#before-start)
    - [Clone authzopa repo](#clone-authzopa-repo)
    - [Reference](#reference)


### Before Start 
- gvm 
- go1.14.6
- istio 1.5.6
  - mixs
  - mixc
- istio/tools
  - mixgen

```bash
# switch go version to 1.14.6
$ gvm use go1.14.6

# create istio 
$ mkdir -p $GOPATH/src/istio.io && cd $GOPATH/src/istio.io

# clone isdio 
$ git clone -b 1.5.6 --depth 1 git@github.com:istio/istio.git

#
$ export MIXER_REPO=$GOPATH/src/istio.io/istio/mixer
$ export ISTIO=$GOPATH/src/istio.io
$ pushd $ISTIO/istio && make mixs && make mixc

# prepare proto
$ go get github.com/gogo/protobuf/proto
$ go get github.com/gogo/protobuf/protoc-gen-gofast
$ go get github.com/gogo/protobuf/protoc-gen-gogoslick
$ go get github.com/gogo/protobuf/protoc-gen-gogo

# if you meet protoc-gen-docs: program not found or is not executable
# you need to clone istio/tools.git and build by yourself
$ cd $GOPATH/src/istio.io && git clone git@github.com:istio/tools.git && cd $GOPATH/src/istio.io/tools/cmd/protoc-gen-docs/
$ make build && mv protoc-gen-docs $GOPATH/bin

$ cd $GOPATH/src/istio.io/istio/mixer/tools/mixgen
$ go build && mv mixgen $GOPATH/bin
```

### Clone authzopa repo

```bash
# git clone authzopa adapter
# fetcher: https://github.com/Gyumeijie/github-files-fetcher 
$ cd $MIXER_REPO/adapter
$ fetcher --url=https://github.com/cage1016/gokit-istio-security/tree/master/authzopa
```


### Reference
- [Mixer Out of Process Adapter Walkthrough · istio/istio Wiki](https://github.com/istio/istio/wiki/Mixer-Out-Of-Process-Adapter-Walkthrough)