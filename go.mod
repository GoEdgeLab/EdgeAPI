module github.com/TeaOSLab/EdgeAPI

go 1.15

replace github.com/TeaOSLab/EdgeCommon => ../EdgeCommon

require (
	github.com/TeaOSLab/EdgeCommon v0.0.0-00010101000000-000000000000
	github.com/go-sql-driver/mysql v1.5.0
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/golang/protobuf v1.4.2
	github.com/iwind/TeaGo v0.0.0-20201020081413-7cf62d6f420f
	github.com/lionsoul2014/ip2region v2.2.0-release+incompatible
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mozillazg/go-pinyin v0.18.0
	github.com/pkg/sftp v1.12.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	google.golang.org/grpc v1.32.0
	google.golang.org/protobuf v1.25.0
)
