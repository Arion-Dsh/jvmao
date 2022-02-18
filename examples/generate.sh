protoc -I proto \
    --go_out=proto \
    --go_opt=paths=source_relative \
    --go-grpc_out=proto \
    --go-grpc_opt=paths=source_relative \
    proto/*.proto

protoc -I proto \
    --js_out=import_style=commonjs:static \
    --grpc-web_out=import_style=commonjs,mode=grpcweb:static \
    proto/*.proto



(cd ./static && ./generate.sh)
