// Copyright 2024 GoEdge CDN goedge.cdn@gmail.com. All rights reserved. Official site: https://goedge.cn .

package instances

type Options struct {
	IsTesting bool
	Verbose   bool
	Cacheable bool

	WorkDir string
	SrcDir  string
	DB      struct {
		Host     string
		Port     int
		Username string
		Password string
		Name     string
	}
	AdminNode struct {
		Port int
	}
	APINode struct {
		HTTPPort     int
		RestHTTPPort int
	}
	Node struct {
		HTTPPort int
	}
	UserNode struct {
		HTTPPort int
	}
}
