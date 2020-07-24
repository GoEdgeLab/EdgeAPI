#!/usr/bin/env bash

ADMIN_PROJECT="../../EdgeAdmin"

protoc --go_out=plugins=grpc:../internal/rpc --proto_path=../internal/rpc/protos ../internal/rpc/protos/*.proto

#admin
cp ../internal/rpc/protos/service_admin.proto ${ADMIN_PROJECT}/internal/rpc/protos/
cp ../internal/rpc/pb/service_admin.pb.go ${ADMIN_PROJECT}/internal/rpc/pb/
cp ../internal/rpc/pb/model_*.go ${ADMIN_PROJECT}/internal/rpc/pb/

