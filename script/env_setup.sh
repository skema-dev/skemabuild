#!/bin/bash

function install_grpc_protos() {
    protos_dir="$HOME/.skema-dev/protos"
    mkdir -p $protos_dir
    git clone https://github.com/googleapis/api-common-protos.git $protos_dir/google-api-common-protos --depth=1
    git clone https://github.com/envoyproxy/protoc-gen-validate.git $protos_dir/envoy-grpc-validate --depth=1
    git clone https://github.com/grpc-ecosystem/grpc-gateway.git $protos_dir/grpc-gateway --depth=1

    protobuf_dir="$HOME/.skema-dev/protobuf"
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
}

rm -rf ~/.skema-dev
install_grpc_protos
exit 0