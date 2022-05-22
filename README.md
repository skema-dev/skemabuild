# skema-tool
dev tools for skema.  
Generate stubs and service code automatically from a single protobuf file.  

## Quick Start
First, make sure you have golang installed.  
Then setup necessary protocol buffers related tools:  
```
sh ./script/setup.sh
```
Now build and test(The following example is using my own sample repo. Change with your own account and repo)
```
cd cmd/cli
go build -ldflags "-X skema-tool/internal/auth.ClientID=<github app clientid>" -o st
./st api init --package=pack1 --service=hello
./st api create --go_option github.com/likezhang-public/newst/test2/com.test/grpc-go --input ./Hello1.proto -o ./stub-test
./st api publish --input=./stub-test --url  https://github.com/likezhang-public/newst/test2 --version=v0.0.2
go get github.com/likezhang-public/newst/test2/com.test/grpc-go@v0.0.2
```
This will generate a `hello.proto` template for protocol buffers definition  
