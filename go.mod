module github.com/TeaOSLab/EdgeAPI

go 1.15

replace github.com/TeaOSLab/EdgeCommon => ../EdgeCommon

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/TeaOSLab/EdgeCommon v0.0.0-00010101000000-000000000000
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1183
	github.com/andybalholm/brotli v1.0.4
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/go-acme/lego/v4 v4.5.2
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.5.2
	github.com/iwind/TeaGo v0.0.0-20211026123858-7de7a21cad24
	github.com/iwind/gosock v0.0.0-20210722083328-12b2d66abec3
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mozillazg/go-pinyin v0.18.0
	github.com/pkg/sftp v1.12.0
	github.com/shirou/gopsutil v3.21.5+incompatible
	github.com/tklauser/go-sysconf v0.3.6 // indirect
	golang.org/x/crypto v0.0.0-20220214200702-86341886e292
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9
	google.golang.org/grpc v1.44.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)
