// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.
//go:build !plus
// +build !plus

package nodes

import "google.golang.org/grpc"

func APINodeServicesRegister(node *APINode, server *grpc.Server) {
}
