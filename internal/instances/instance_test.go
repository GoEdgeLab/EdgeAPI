// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package instances_test

import (
	"github.com/TeaOSLab/EdgeAPI/internal/instances"
	"github.com/iwind/TeaGo/Tea"
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

var instance = instances.NewInstance(instances.Options{
	Cacheable: true,
	WorkDir:   Tea.Root + "/standalone-instance",
	SrcDir:    Tea.Root + "/standalone-instance/src",
	DB: struct {
		Host     string
		Port     int
		Username string
		Password string
		Name     string
	}{
		Host:     "127.0.0.1",
		Port:     3306,
		Username: "root",
		Password: "123456",
		Name:     "edges2",
	},
	AdminNode: struct {
		Port int
	}{
		Port: 7788,
	},
	APINode: struct {
		HTTPPort     int
		RestHTTPPort int
	}{
		HTTPPort:     8001,
		RestHTTPPort: 8002,
	},
	Node: struct{ HTTPPort int }{
		HTTPPort: 8080,
	},
	UserNode: struct {
		HTTPPort int
	}{
		HTTPPort: 7799,
	},
})

func TestInstanceSetupAll(t *testing.T) {
	err := instance.SetupAll()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstance_SetupDB(t *testing.T) {
	err := instance.SetupDB()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstance_SetupAdminNode(t *testing.T) {
	err := instance.SetupAdminNode()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstance_SetupAPINode(t *testing.T) {
	err := instance.SetupAPINode()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstance_SetupUserNode(t *testing.T) {
	err := instance.SetupUserNode()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstance_SetupNode(t *testing.T) {
	err := instance.SetupNode()
	if err != nil {
		t.Fatal(err)
	}
}

func TestInstance_Clean(t *testing.T) {
	err := instance.Clean()
	if err != nil {
		t.Fatal(err)
	}
}
