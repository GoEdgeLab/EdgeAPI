module github.com/TeaOSLab/EdgeAPI

go 1.15

replace github.com/TeaOSLab/EdgeCommon => ../EdgeCommon

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/TeaOSLab/EdgeCommon v0.0.0-00010101000000-000000000000
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.641
	github.com/go-acme/lego/v4 v4.1.2
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/iwind/TeaGo v0.0.0-20210106152225-413a5aba30aa // indirect
	github.com/lionsoul2014/ip2region v2.2.0-release+incompatible
	github.com/mozillazg/go-pinyin v0.18.0
	github.com/pkg/sftp v1.12.0
	github.com/shirou/gopsutil v2.20.9+incompatible
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
