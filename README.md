# skema-tool
dev tools for skema.  
Generate stubs and service code automatically from a single protobuf file.  

## Quick Start
First, make sure you have golang installed.  
Then setup necessary protocol buffers related tools:  
```
sh ./script/setup.sh
```
Now build and test  
```
cd cmd/cli
go build -ldflags "-X skema-tool/internal/auth.ClientID=<github app clientid>" -o st
./st api init --package=pack1 --service=hello
```
This will generate a `hello.proto` template for protocol buffers definition  

