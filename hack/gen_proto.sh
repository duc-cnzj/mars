#!/usr/bin/env bash

PROTO_FILES=$(find ../ \(  -name "*.proto"  -not -path "*/common/*" \) | sort)
for i in ${PROTO_FILES}; do
protoc \
    -I ../server/api/protos \
    -I ../server/websocket/protos \
    -I ../server/common/protos \
    --go_out=../pkg \
    --go-grpc_out=../pkg \
    --go_opt=paths=source_relative \
    --go-grpc_opt=paths=source_relative \
    --openapiv2_out=../doc \
    --openapiv2_opt logtostderr=true \
    --openapiv2_opt json_names_for_fields=false \
    --grpc-gateway_out=../pkg \
    --grpc-gateway_opt logtostderr=true \
    --grpc-gateway_opt paths=source_relative \
    --grpc-gateway_opt generate_unbound_methods=true \
    $i
done

pbjs -t static-module -o ../frontend/src/api/compiled.js -w es6  ../server/api/protos/**/*.proto --keep-case --no-create --no-encode --no-decode --no-verify --no-convert --no-delimited
pbts -o ../frontend/src/api/compiled.d.ts ../frontend/src/api/compiled.js --keep-case

#   --no-create      Does not generate create functions used for reflection compatibility.
#   --no-encode      Does not generate encode functions.
#   --no-decode      Does not generate decode functions.
#   --no-verify      Does not generate verify functions.
#   --no-convert     Does not generate convert functions like from/toObject
#   --no-delimited   Does not generate delimited encode/decode functions.
#   --no-beautify    Does not beautify generated code.
#   --no-comments    Does not output any JSDoc comments.
#   --no-service     Does not output service classes.