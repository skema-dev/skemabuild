#!/bin/bash

function command_exists() {
	command -v "$@" > /dev/null 2>&1
}

function add_env_sh() {
    result=$(cat $1 | grep "SKEMA_HOME")
    if [[ "$result" != "" ]]
    then
        echo ""
    else
        echo 'export SKEMA_HOME="${HOME}/.skema"' >> $1
        echo 'export PATH="${PATH}:${SKEMA_HOME}"' >> $1
    fi
}

# common function to set PATH for different OS. zshrc for macos, bashrc and bash_profile for linux
function set_environments() {
    echo "update environment profile"
     # set bash and zsh
     if [[ -f ~/.zshrc ]]; then
         add_env_sh ~/.zshrc
     fi
     if [[ -f ~/.bashrc ]]; then
         add_env_sh ~/.bashrc
     fi
     if [[ -f ~/.bash_profile ]]; then
         add_env_sh ~/.bash_profile
     fi
 }

function install_grpc_protos() {
    protos_dir="$HOME/.skema/protos"
    mkdir -p $protos_dir
    git clone https://github.com/googleapis/api-common-protos.git $protos_dir/google-api-common-protos --depth=1
    git clone https://github.com/envoyproxy/protoc-gen-validate.git $protos_dir/envoy-grpc-validate --depth=1
    git clone https://github.com/grpc-ecosystem/grpc-gateway.git $protos_dir/grpc-gateway --depth=1

    protobuf_dir="$HOME/.skema/protobuf"
    mkdir -p $protobuf_dir
    git -C $protobuf_dir init -b main
    git -C $protobuf_dir remote add origin https://github.com/protocolbuffers/protobuf.git
    git -C $protobuf_dir config core.sparsecheckout true
    echo "/src" >> $protobuf_dir/.git/info/sparse-checkout
    git -C $protobuf_dir pull --depth 1 --ff-only origin main
    cp -r $protobuf_dir/src $protos_dir/protobuf
    rm -rf $protobuf_dir/src

    # install protoc
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        apt install -y protobuf-compiler
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install protobuf
    fi

    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.10.0
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.10.0
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0

    go install github.com/envoyproxy/protoc-gen-validate@v0.6.7
}

function install_kind() {
    echo "try installing kind for local kubernetes cluster"
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-amd64
        chmod +x ./kind
        mv ./kind /usr/local/bin/kind
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install kind
    else
        echo "Install Kind Using go install"
        go install sigs.k8s.io/kind@v0.14.0
    fi    
}

function install_docker() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl -fsSL https://get.docker.com/ | bash
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install --cask docker
    else
        echo "Please install docker manually"
    fi
}

function install_docker_compose() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl -L "https://github.com/docker/compose/releases/download/v2.6.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install docker-compose
    else
        echo "Please install docker-compose manually"
    fi
}

function install_docker_tools() {
    echo "Install docker and docker-compose"
    docker_versioninfo=""
    docker_versioninfo=$(docker --version | head -n 1)
    if [ -z "$docker_versioninfo" ]; then
        echo docker not exists;
        install_docker
    else
        echo docker already exists, version is: $docker_versioninfo ;
    fi

    if command_exists docker-compose; then
        echo docker-compose already exists;
    else
        echo docker-compose not exists;
        install_docker_compose
    fi


    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.14.0/kind-linux-amd64
        chmod +x ./kind
        mv ./kind /usr/local/bin/kind
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install kind
    else
        echo "Install Kind Using go install"
        go install sigs.k8s.io/kind@v0.14.0
    fi    
}



rm -rf ~/.skema
install_grpc_protos
install_docker_tools
install kind
set_environments

cmd="go install github.com/skema-dev/skemabuild/cmd/skbuild@latest"
echo $cmd
eval $cmd

exit 0
