#!/bin/bash

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

rm -rf ~/.skema
install_grpc_protos
set_environments

go install github.com/skema-dev/skema-tool/cmd/st

exit 0