# gRPC-Gateway

## TL;DR

สร้างไฟล์ `*.pb.go`, `*_grpc.pb.go` และ `*.pb.gw.go` จาก `helloworld.proto`  
โดยใช้ docker image `thethingsindustries/protoc` ช่วย compile

``` bash
# Help & flags
docker run --rm thethingsindustries/protoc:3.1.33 --help

# Generate message stub
# - proto/helloworld/helloworld.pb.go
docker run --rm -v d/workdir/grpc/grpc_gateway:/mnt/src -w /mnt/src thethingsindustries/protoc:3.1.33 \
    --go_out=/mnt/src/proto \
    --proto_path=/mnt/src/proto/helloworld \
    /mnt/src/proto/helloworld/helloworld.proto

# Generate grpc services stub
# - proto/helloworld/helloworld_grpc.pb.go
docker run --rm -v d/workdir/grpc/grpc_gateway:/mnt/src -w /mnt/src thethingsindustries/protoc:3.1.33 \
    --go-grpc_out=/mnt/src/proto \
    --proto_path=/mnt/src/proto/helloworld \
    /mnt/src/proto/helloworld/helloworld.proto

# Generate grpc-gateway stub
# - proto/helloworld/helloworld.pb.gw.go
# Note: read about --grpc-gateway_opt at: https://github.com/grpc-ecosystem/grpc-gateway/
docker run --rm -v d/workdir/grpc/grpc_gateway:/mnt/src -w /mnt/src thethingsindustries/protoc:3.1.33 \
    --grpc-gateway_opt='logtostderr=true' \
    --grpc-gateway_opt='paths=source_relative' \
    --grpc-gateway_opt='generate_unbound_methods=true' \
    --grpc-gateway_out=/mnt/src/proto/helloworld \
    --proto_path=/mnt/src/proto/helloworld \
    /mnt/src/proto/helloworld/helloworld.proto

# Or onetime generate all helloworld stubs(messages, grpc-services and grpc-gateway)
# docker run --rm -v d/workdir/grpc/grpc_gateway:/mnt/src -w /mnt/src thethingsindustries/protoc:3.1.33 \
#     --go_out=/mnt/src/proto \
#     --go-grpc_out=/mnt/src/proto \
#     --grpc-gateway_opt='logtostderr=true' \
#     --grpc-gateway_opt='paths=source_relative' \
#     --grpc-gateway_opt='generate_unbound_methods=true' \
#     --grpc-gateway_out=/mnt/src/proto/helloworld \
#     --proto_path=/mnt/src/proto/helloworld \
#     /mnt/src/proto/helloworld/helloworld.proto
```

สร้างไฟล์ `*.pb.go`, `*_grpc.pb.go` และ `*.pb.gw.go` จาก `pingpong.proto` ซึ่งเขียน HTTP semantics spec(method, path, etc.) เอาไว้ด้วย

``` bash
# Onetime generate all pingpong stubs(messages, grpc-services and grpc-gateway)
docker run --rm -v d/workdir/grpc/grpc_gateway:/mnt/src -w /mnt/src thethingsindustries/protoc:3.1.33 \
    --go_out=/mnt/src \
    --go-grpc_out=/mnt/src \
    --grpc-gateway_opt='logtostderr=true' \
    --grpc-gateway_opt='paths=source_relative' \
    --grpc-gateway_opt='generate_unbound_methods=false' \
    --grpc-gateway_out=/mnt/src/proto/pingpong \
    --proto_path=/mnt/src/proto/pingpong \
    /mnt/src/proto/pingpong/pingpong.proto
```

สร้างไฟล์ OpenAPI configs(`helloworld.swagger.json` และ `pingpong.swagger.json`) ด้วย protoc
เพื่อนำไปใช้งานใน [swagger](https://swagger.io/tools/swagger-editor/)

``` bash
go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc

cd <project_root>/proto 

# Generate helloworld/helloworld.swagger.json *generate_unbound_methods=true*
protoc -I . \
    --openapiv2_out . \
    --openapiv2_opt logtostderr=true \
    --openapiv2_opt generate_unbound_methods=true \
    --proto_path=./helloworld \
    ./helloworld/helloworld.proto

# Generate pingpong/pingpong.swagger.json *generate_unbound_methods=false*
protoc -I . \
    --openapiv2_out . \
    --openapiv2_opt logtostderr=true \
    --openapiv2_opt generate_unbound_methods=false \
    --proto_path=./pingpong \
    ./pingpong/pingpong.proto    
```

## Testing the gRPC-Gateway

Start the server

``` bash
go run main.go
```

cURL to send HTTP requests

``` bash
$ curl -X POST -k http://localhost:8090/helloworld.Greeter/SayHello -d '{"name": "hello"}'
{"message":"hello  world"}

$ curl -X POST -k http://localhost:8090/v1/example/pingpong -d '{"source": "127.0.0.1", "destination": "127.0.0.2"}'
{"result":"ok"}
```

## Customise the HTTP semantics (method, path, etc.)

Requirements

``` bash
# for protoc
go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

การตั้งค่า path endpoint มี 3 รูปแบบ (https://github.com/grpc-ecosystem/grpc-gateway#usage หัวข้อ 4)  

1. ใช้ default mapping HTTP semantics (method, path, etc.) คือใช้ไฟล์ `.proto` แบบเดิมๆที่เขียนแค่ protobuf/grpc-service เลย ข้อเสียคือ custom ให้เป็นค่าอื่นๆไม่ได้ เราสามารถ generate stub file ด้วย `protoc` ได้ดังนี้

    ``` bash
    cd proto/
    protoc -I . --grpc-gateway_out ./ \
        --grpc-gateway_opt logtostderr=true \
        --grpc-gateway_opt paths=source_relative \
        --grpc-gateway_opt generate_unbound_methods=true \
        ./helloworld/helloworld.proto
    ```

2. เพิ่ม information ลงไป `.proto` เพื่อระบุ HTTP semantics (method, path, etc.) ลงไปเพิ่มเติม  
วิธีนี้เราต้องเพิ่มข้อมูลลงไปใน service ตัวอย่างเช่น

    ```  protobuf
    ...
    import "google/api/annotations.proto";
    ...

    service PingPongService {
        rpc Pingpong(Ping) returns (Pong) {
            option (google.api.http) = {
                    post: "/v1/demo/pingpong"
                    body: "*"
                };
        }
    }
    ```

    สามารถดูตัวอย่างการเขียนเพิ่มเติมได้ที่ [link](https://github.com/grpc-ecosystem/grpc-gateway/blob/master/examples/internal/proto/examplepb/a_bit_of_everything.proto)

    หากใช้งาน `protoc` ในการ generate stubs เราจะต้องไป copy [google apis dependencies](https://github.com/googleapis/googleapis) มาใส่โปรเจคเพิ่มเติมเพื่อใช้ในขั้นตอนการ compile ไฟล์ `*.proto`

    ``` files
    google/api/annotations.proto
    google/api/field_behaviour.proto
    google/api/http.proto
    google/api/httpbody.proto
    ```

    ``` tree
    proto/
    ├── google/
    │   └── api/
    │       ├── annotations.proto
    │       ├── field_behaviour.proto
    │       ├── http.proto
    │       └── httpbody.proto
    ├── helloworld/
    │   └── helloworld.proto
    │
    ...
    │
    └── pingpong/
        └── pingpong.proto
    ```

    Note: ถ้าไม่ใช้ protoc อีกวิธีคือใช้ [buf](https://github.com/bufbuild/buf) ซึ่งใช้วิธีเขียน file config และระบุ dependencies แทน แต่ตอนกำลังเขียนมันยังใหม่เกินไป เลยไม่เลือกใช้  

    สามารถ compile เพื่อสร้าง *.gw.pb.go ได้ดังนี้

    ``` bash
    cd proto/
    protoc -I . --grpc-gateway_out ./ \
        --grpc-gateway_opt logtostderr=true \
        --grpc-gateway_opt paths=source_relative \
        --grpc-gateway_opt generate_unbound_methods=false \
        ./pingpong/pingpong.proto
    ```

    ตัวอย่างการ compile gRPC-Gateway สำหรับ `pingpong.proto`

    ``` bash
    # Onetime generate all stubs(messages, grpc-services and grpc-gateway)
    docker run --rm -v d/workdir/grpc/grpc_gateway:/mnt/src -w /mnt/src thethingsindustries/protoc:3.1.33 \
        --go_out=/mnt/src \
        --go-grpc_out=/mnt/src \
        --grpc-gateway_opt='logtostderr=true' \
        --grpc-gateway_opt='paths=source_relative' \
        --grpc-gateway_opt='generate_unbound_methods=false' \
        --grpc-gateway_out=/mnt/src/proto/pingpong \
        --proto_path=/mnt/src/proto/pingpong \
        /mnt/src/proto/pingpong/pingpong.proto
    ```

3. ในกรณีที่ไม่ต้องการแก้ไขหรือไม่สามารถเขียน spec ลงไปใน `.proto` ได้เราสามารถเขียนเป็นไฟล์ config .yaml เพื่อใช้ระบุ spec ของ HTTP semantics (method, path, etc.) ด้วยการระบุไฟล์ config เข้าไปด้วย `--grpc-gateway_opt grpc_api_configuration=path/to/config.yaml` นอกจากนั้นให้ใส่ `--grpc-gateway_opt standalone=true` เข้าไปด้วยเพื่อให้การอ้างอิง types ของไฟล์ `xxx.pb.gw.go` ไปยัง external source ได้อย่างถูกต้อง

    ``` bash
    cd proto/
    protoc -I . --grpc-gateway_out ./ \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt paths=source_relative \
    --grpc-gateway_opt grpc_api_configuration=path/to/config.yaml \
    --grpc-gateway_opt standalone=true \
    your/service/v1/your_service.proto
    ```

    สำหรับการเขียน config.yaml สามารถอ่านได้ที่ [link](https://cloud.google.com/endpoints/docs/grpc/grpc-service-config)

## Generate OpenAPI definitions using protoc-gen-openapiv2

เนื่องจาก `thethingsindustries/protoc` image ยังไม่มีการ support `protoc-gen-openapiv2`  
เลยต้อง compile ผ่าน protoc ตรงๆด้วยคำสั่ง

``` bash
# Install protoc-gen-openapiv2 plugin
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
```

helloworld.proto ใช้ default mapping HTTP semantics(ไม่ได้ ้http spec เพิ่มลงไปใน .proto เลย)  
ดังนั้นการสร้าง `helloworld.swagger.json` จะต้องใส่ flag `generate_unbound_methods=true` เข้าไปด้วย

``` bash
cd <project_root>/proto 

# Generate helloworld/helloworld.swagger.json
protoc -I . \
    --openapiv2_out . \
    --openapiv2_opt logtostderr=true \
    --openapiv2_opt generate_unbound_methods=true \
    --proto_path=./helloworld \
    ./helloworld/helloworld.proto
```

สำหรับ `pingpong.proto` ที่ใช้วิธีการเขียน spec ของ HTTP semantics ลงไปในไฟล์ .proto โดยตรง  
การสร้าง `pingpong.swagger.json` จึงไม่ต้องกำหนด `generate_unbound_methods` ลงไปก็ได้  
หรือถ้าจะใส่ก็ใส่เป็นค่า `generate_unbound_methods=false`

``` bash
cd <project_root>/proto 

# Generate pingpong/pingpong.swagger.json
protoc -I . \
    --openapiv2_out . \
    --openapiv2_opt logtostderr=true \
    --proto_path=./pingpong \
    ./pingpong/pingpong.proto
```

เราสามารถ copy เนื้อหาในไฟล์ `helloworld.swagger.json` หรือ `pingpong.swagger.json` ไปใส่ใน Live Demo ของ swagger ได้ที่  
https://swagger.io/tools/swagger-editor/  หน้าเว็บจะ auto convert จาก JSON ให้เป็น YAML ให้เราโดยอัตโนมัติ

## References

- [gRPC.io](https://grpc.io/)
- [BloomRPC: gRPC Postman-like](https://github.com/uw-labs/bloomrpc)
- [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway/)
- [gROC-Gateway tutorial](https://grpc-ecosystem.github.io/grpc-gateway/docs/tutorials/adding_annotations/)
- [How gRPC services map to the JSON request and response](https://github.com/grpc-ecosystem/grpc-gateway/blob/master/examples/internal/proto/examplepb/a_bit_of_everything.proto)
- [The 'mustEmbedUnimplemented' issue](https://github.com/grpc/grpc-go/issues/3794#issuecomment-720599532)
- [buf: Stubs generator](https://github.com/bufbuild/buf)
- [buf: Installation](https://docs.buf.build/installation/)
- [docker-protobuf](https://github.com/TheThingsIndustries/docker-protobuf)
- [HTTP Graceful Shutdown](https://medium.com/honestbee-tw-engineer/gracefully-shutdown-in-go-http-server-5f5e6b83da5a)
- [Online swagger-editor](https://swagger.io/tools/swagger-editor/)
- [Example gRPC gateway protobuf file: google pubsub](https://github.com/googleapis/googleapis/blob/master/google/pubsub/v1/pubsub.proto)
- [Example gRPC gateway protobuf file: google spanner](https://github.com/googleapis/googleapis/blob/master/google/spanner/v1/spanner.proto)
