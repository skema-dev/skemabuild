
# 1. Init and Publish Protobuf API
Now run and test(The following example is using my own sample repo. Change with your own account and repo)
```shell
# create an initial protocol buffers file
skbuild api init --package=com.pack1 --service=hello

# create stubs for grpc-go and openapi, using protobuf file Hello1.proto, and output to ./stub-test
skbuild api create --go_option github.com/likezhang-public/newst/test001/com.pack1/grpc-go --input ./hello.proto -o ./stub-test

# after stubs are created, upload to github and set the version tag
# the "url" argument is for verifying the go_package option defined above
skbuild api publish --stub=./stub-test --url  https://github.com/likezhang-public/newst/test001 --version=v0.0.1

# verify the package is downloadable
skbuild get github.com/likezhang-public/newst/test001/com.pack1/grpc-go@v0.0.1
```

You can also skip creating local stubs and publish to github directly:
```shell
skbuild api init --package=com.pack1 --service=hello

# here we combine the two steps (create and publish) into a single process to simplify the workflow
skbuild api publish --proto=./hello.proto --url  https://github.com/likezhang-public/newst/test2 --version=v0.0.2

go get github.com/likezhang-public/newst/test2/com.pack1/grpc-go@v0.0.2
```

How about publishing to non-github repository? How about gitlab/bitbucket or private git hosting?  
Check out following sections for details. It's almost the same process :)
<hr/>

# 2. Generate gRPC Service Code
Now it's the final step: generate runnable gRPC service code automatically!  
Just use the following example to specify the protobuf file you published online
```shell
skbuild service create --proto https://github.com/likezhang-public/newst/blob/main/test2/com.pack1/grpc-go/Hello.proto
```
Then we'll have the following code layout in current path. (you can specify output path by using --output)
.
├── cmd  
│     └── hello1service  
│         ├── config  
│         │   └── grpc.yaml  
│         ├── main.go  
│         └── service  
│             └── service.go  
├── go.mod  
└── internal  
└── README.md
<hr/>

# 3. Run the service code
To test it out, run `go mod tidy` first to install dependencies, then go to `cmd/hello1service` and `go run .`
```shell
go mod tidy
cd cmd/helloservice
go run .
```
You'll see the launching information of grpc service and the http endpoint gateway:
```
2022-06-07T00:48:32.460+0800	INFO	logging/logging.go:96	using local config from ./config/grpc.yaml
2022-06-07T00:48:32.460+0800	INFO	logging/logging.go:66	service port	{"gprc": 9991, "http": 9992}
2022-06-07T00:48:32.461+0800	INFO	logging/logging.go:96	gateway path is set to /
2022-06-07T00:48:32.461+0800	INFO	logging/logging.go:66	logging initialized:	{"level": "debug", "encoding": "console"}
2022-06-07T00:48:32.461+0800	INFO	logging/logging.go:96	Serving gRPC start...
```
To further understand how the code is designed, Please checkout our [Skema-Go Framework](https://github.com/skema-dev/skema-go)
<hr/>

# 4. Support for other GIT Servers
SkemaBuild is also capable of publishing stubs to your private git server. Underlying the hood, it's using `git-go` package to simulate git client and pushing files to repo.  
Comparing with publishing on github, you'll have to do three things differently:
- You need to have ~/.ssh/id_rsa file and authenticated in your git server (e.g. gitlab or others)
- Clone the git repo on your local path
- Specify the go_package option explicitly when creating stubs

That's it. The whole workflow is the same as previously. Let's use gitlab as an example:
```
# clone the git repo
git clone git@gitlab.com:likexx/test1.git
cd test1

# init protobuf file
skbuild api init --package=org.test --service=Hello3

# generate stubs and output to ./temp, using user specified go_option
skbuild api create -i ./Hello3.proto -o ./temp --go_option "gitlab.com/likexx/test1/stub/org.test/grpc-go"

# publish stubs from ./temp to the repo (assuming the command is executed at the repository root).
# "stub" is the user specified path in the repo
skbuild api publish -s ./temp -u stub --version v1.0.2

# create grpc service from the raw protobuf file link
# you may need to run `skbuild auth -t github` in order to use the code template on github repo
skbuild service create -p https://gitlab.com/likexx/test1/-/raw/main/stub/org.test/grpc-go/Hello3.proto

go mod tidy
cd cmd
cd hello3service
go run .
```

# 5. Use Username and Password for Git Http Auth
Some users might clone repo using https instead of ssh. No problem! Just specify your username and password when publishing protobuf stubs:
```
skbuild api publish -s ./temp -u stub --version v1.0.1 --username=<your git username> --password=<your password>
```

# 6. Cloud Native Local Dev Environment!!
After the code is generated (using the standard template), you'll have a "env" folder, which contains script for kubernetes cluster setup (via kind), and Dockerfile/docker-compose/k8s deploy config/deploy scripts/etc.

Yes, we generate everything you need to pack your code into docker image, and play around with either docker-compose or local kubernetes cluster (the install script will add Kind on your machine).

Not just generating the scripts/configurations, we also simplified and standardlized the cluster/service setup/deployment in our skbuild tool. You can simply run the following commands:
```
# create a kind cluster
# save the output token for dashboard access
skbuild dev cluster create

# build docker image for the service
skbuild dev service imagebuild -n <your servicename>

# deploy the service on the cluster
skbuild dev service create -n hello3service

# test the service
curl http://localhost:30002/api/healthcheck

# you can check the k8s dashboard at:
# http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/#/login

# delete the service on cluster
skbuild dev service delete -n hello3service

# delete the cluster
skbuild dev cluster delete
```
