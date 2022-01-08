#!/usr/bin/env bash

PROTO_FILES=$(find ../ \( -name "*.proto"  -not -path "*/common/*" -not -path "*/node_modules/*" \) | sort)
for i in ${PROTO_FILES}; do
protoc \
    -I ../internal/grpc/protos \
    -I ../internal/grpc/common/protos \
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
    --validate_out=lang=go,paths=source_relative:../pkg \
    $i
done

pbjs -t static-module -o ../frontend/src/api/compiled.js -w es6  ../internal/grpc/protos/**/*.proto \
  --keep-case \
  --no-verify \
  --no-convert \
  --no-create \
  --force-number \
  --force-message \
  --no-delimited
#  --no-encode \
#  --no-decode \

pbts -o ../frontend/src/api/compiled.d.ts ../frontend/src/api/compiled.js --keep-case

# https://github.com/protobufjs/protobuf.js/blob/master/cli/README.md#reflection-vs-static-code
#  Static targets only:
#
#  --no-create      Does not generate create functions used for reflection compatibility.
#  --no-encode      Does not generate encode functions.
#  --no-decode      Does not generate decode functions.
#  --no-verify      Does not generate verify functions.
#  --no-convert     Does not generate convert functions like from/toObject
#  --no-delimited   Does not generate delimited encode/decode functions.
#  --no-beautify    Does not beautify generated code.
#  --no-comments    Does not output any JSDoc comments.
#  --no-service     Does not output service classes.
#
#  --force-long     Enforces the use of 'Long' for s-/u-/int64 and s-/fixed64 fields.
#  --force-number   Enforces the use of 'number' for s-/u-/int64 and s-/fixed64 fields.
#  --force-message  Enforces the use of message instances instead of plain objects.

swagger mixin --ignore-conflicts ../third_party/doc/data/api.json ../doc/**/*.json > ../third_party/doc/data/swagger.json