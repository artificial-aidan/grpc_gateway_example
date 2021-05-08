#!/bin/sh

# Generate message and grpc stubs
# - proto/helloworld/helloworld.pb.go
# - proto/helloworld/helloworld_grpc.pb.go
docker run --rm -v d/workdir/grpc/grpc_gateway:/mnt/src -w /mnt/src thethingsindustries/protoc:3.1.33 \
    --proto_path=/mnt/src/proto/helloworld \
    --go_out=/mnt/src/proto \
    --go-grpc_out=/mnt/src/proto \
    /mnt/src/proto/helloworld/helloworld.proto

# Generate grpc-gateway stubs
# - proto/helloworld/helloworld.pb.gw.go
docker run --rm -v d/workdir/grpc/grpc_gateway:/mnt/src -w /mnt/src thethingsindustries/protoc:3.1.33 \
    --proto_path=/mnt/src/proto/helloworld \
    --grpc-gateway_out=/mnt/src/proto/helloworld \
    --grpc-gateway_opt='logtostderr=true' \
    --grpc-gateway_opt='paths=source_relative' \
    --grpc-gateway_opt='generate_unbound_methods=true' \
    /mnt/src/proto/helloworld/helloworld.proto
