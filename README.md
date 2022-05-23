# skema-tool
dev tools for skema.  
Generate stubs and service code automatically from a single protobuf file.  

## Quick Start
First, make sure you have golang installed.  
Then setup necessary protocol buffers related tools:  
```shell
curl -L https://raw.githubusercontent.com/skema-dev/skema-tool/main/script/setup.sh | bash
```
Now run and test(The following example is using my own sample repo. Change with your own account and repo)
```shell
# craete an initial protocol buffers file
st api init --package=pack1 --service=hello

# create stubs for grpc-go and openapi, using protobuf file Hello1.proto, and output to ./stub-test
st api create --go_option github.com/likezhang-public/newst/test2/com.test/grpc-go --input ./Hello1.proto -o ./stub-test

# after stubs are created, upload to github and set the version tag
st api publish --stub=./stub-test --url  https://github.com/likezhang-public/newst/test2 --version=v0.0.2

# verify the package is downloadable
go get github.com/likezhang-public/newst/test2/com.test/grpc-go@v0.0.2
```
You can also skip creating local stubs and publish to github directly:  
```shell
st api init --package=pack1 --service=hello

st api publish --proto=./hello.proto --url  https://github.com/likezhang-public/newst/test2 --version=v0.0.2

go get github.com/likezhang-public/newst/test2/com.test/grpc-go@v0.0.2
```

