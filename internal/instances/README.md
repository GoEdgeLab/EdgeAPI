# 实例

## 目录结构：
~~~
/opt/cache
/usr/local/goedge/
	edge-admin/
	   configs/
	      api_admin.yaml
	      api_db.yaml
	      server.yaml
	edge-api/
	   configs/api.yaml
	   configs/db.yaml
	api-node/
	   configs/api_node.yaml
	api-user/
	   configs/api_user.yaml
    src/	   
/usr/bin/
  edge-admin -> ...
  edge-api -> ...
  edge-node -> ...
  edge-user -> ...
/usr/local/mysql
~~~

* 其中 `->` 表示软链接。
* `src/` 目录下放置zip格式的待安装压缩包

## 端口
* Admin：7788
* API：8001
* API HTTP：8002
* User: 7799
* Server: 8080

## 数据库
数据库名称为 `edges`