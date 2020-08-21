#!/usr/bin/env bash

ADMIN_PROJECT="../../EdgeAdmin"
NODE_PROJECT="../../EdgeNode"

rm -f ../internal/rpc/pb/*
protoc --go_out=plugins=grpc:../internal/rpc --proto_path=../internal/rpc/protos ../internal/rpc/protos/*.proto

# admin
function pub() {
	cp ../internal/rpc/protos/service_${2}.proto ${1}/internal/rpc/protos/
	cp ../internal/rpc/pb/service_${2}.pb.go ${1}/internal/rpc/pb/
}

pub ${ADMIN_PROJECT} admin
pub ${ADMIN_PROJECT} node
pub ${ADMIN_PROJECT} node_cluster
pub ${ADMIN_PROJECT} node_grant
pub ${ADMIN_PROJECT} server

cp ../internal/rpc/pb/model_*.go ${ADMIN_PROJECT}/internal/rpc/pb/

# node
pub ${NODE_PROJECT} node

cp ../internal/rpc/pb/model_node.pb.go ${NODE_PROJECT}/internal/rpc/pb/
cp ../internal/rpc/pb/model_node_login.pb.go ${NODE_PROJECT}/internal/rpc/pb/
cp ../internal/rpc/pb/model_node_grant.pb.go ${NODE_PROJECT}/internal/rpc/pb/
cp ../internal/rpc/pb/model_node_cluster.pb.go ${NODE_PROJECT}/internal/rpc/pb/