// generated
package sqls

var SQL_full = "CREATE TABLE `edgeAPINodes` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `clusterId` int(11) unsigned DEFAULT '0' COMMENT '专用集群ID',\n" +
	"  `uniqueId` varchar(32) DEFAULT NULL COMMENT '唯一ID',\n" +
	"  `secret` varchar(32) DEFAULT NULL COMMENT '密钥',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  `http` json DEFAULT NULL COMMENT '监听的HTTP配置',\n" +
	"  `https` json DEFAULT NULL COMMENT '监听的HTTPS配置',\n" +
	"  `accessAddrs` json DEFAULT NULL COMMENT '外部访问地址',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `weight` int(11) unsigned DEFAULT '0' COMMENT '权重',\n" +
	"  `status` json DEFAULT NULL COMMENT '运行状态',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  UNIQUE KEY `uniqueId` (`uniqueId`) USING BTREE\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API节点';\n" +
	"CREATE TABLE `edgeAPITokens` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `nodeId` varchar(32) DEFAULT NULL COMMENT '节点ID',\n" +
	"  `secret` varchar(255) DEFAULT NULL COMMENT '节点密钥',\n" +
	"  `role` varchar(64) DEFAULT NULL COMMENT '节点角色',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `nodeId` (`nodeId`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API令牌管理';\n" +
	"CREATE TABLE `edgeAdmins` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `username` varchar(64) DEFAULT NULL COMMENT '用户名',\n" +
	"  `password` varchar(32) DEFAULT NULL COMMENT '密码',\n" +
	"  `fullname` varchar(64) DEFAULT NULL COMMENT '全名',\n" +
	"  `isSuper` tinyint(1) unsigned DEFAULT '0' COMMENT '是否为超级管理员',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `updatedAt` bigint(11) unsigned DEFAULT '0' COMMENT '修改时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员';\n" +
	"CREATE TABLE `edgeDBNodes` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `role` varchar(255) DEFAULT NULL COMMENT '数据库角色',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  `host` varchar(255) DEFAULT NULL COMMENT '主机',\n" +
	"  `port` int(11) unsigned DEFAULT '0' COMMENT '端口',\n" +
	"  `database` varchar(255) DEFAULT NULL COMMENT '数据库名称',\n" +
	"  `username` varchar(255) DEFAULT NULL COMMENT '用户名',\n" +
	"  `password` varchar(255) DEFAULT NULL COMMENT '密码',\n" +
	"  `charset` varchar(255) DEFAULT NULL COMMENT '通讯字符集',\n" +
	"  `connTimeout` int(11) unsigned DEFAULT '0' COMMENT '连接超时时间（秒）',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `weight` int(11) unsigned DEFAULT '0' COMMENT '权重',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库节点';\n" +
	"CREATE TABLE `edgeFileChunks` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `fileId` int(11) unsigned DEFAULT '0' COMMENT '文件ID',\n" +
	"  `data` longblob COMMENT '分块内容',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `fileId` (`fileId`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件片段';\n" +
	"CREATE TABLE `edgeFiles` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `description` varchar(255) DEFAULT NULL COMMENT '文件描述',\n" +
	"  `filename` varchar(255) DEFAULT NULL COMMENT '文件名',\n" +
	"  `size` int(11) unsigned DEFAULT '0' COMMENT '文件尺寸',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  `type` varchar(64) DEFAULT '' COMMENT '类型',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `type` (`type`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件管理';\n" +
	"CREATE TABLE `edgeHTTPAccessLogPolicies` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `type` varchar(255) DEFAULT NULL COMMENT '存储类型',\n" +
	"  `options` json DEFAULT NULL COMMENT '存储选项',\n" +
	"  `conds` json DEFAULT NULL COMMENT '请求条件',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='访问日志策略';\n" +
	"CREATE TABLE `edgeHTTPAccessLogs` (\n" +
	"  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `serverId` int(11) unsigned DEFAULT '0' COMMENT '服务ID',\n" +
	"  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n" +
	"  `status` int(3) unsigned DEFAULT '0' COMMENT '状态码',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `content` json DEFAULT NULL COMMENT '日志内容',\n" +
	"  `requestId` varchar(128) DEFAULT NULL COMMENT '请求ID',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `serverId` (`serverId`),\n" +
	"  KEY `nodeId` (`nodeId`),\n" +
	"  KEY `serverId_status` (`serverId`,`status`),\n" +
	"  KEY `requestId` (`requestId`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;\n" +
	"CREATE TABLE `edgeHTTPAccessLogs_20201010` (\n" +
	"  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `serverId` int(11) unsigned DEFAULT '0' COMMENT '服务ID',\n" +
	"  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n" +
	"  `status` int(3) unsigned DEFAULT '0' COMMENT '状态码',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `content` json DEFAULT NULL COMMENT '日志内容',\n" +
	"  `day` varchar(8) DEFAULT NULL COMMENT '日期Ymd',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `serverId` (`serverId`),\n" +
	"  KEY `nodeId` (`nodeId`),\n" +
	"  KEY `serverId_status` (`serverId`,`status`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;\n" +
	"CREATE TABLE `edgeHTTPCachePolicies` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `capacity` json DEFAULT NULL COMMENT '容量数据',\n" +
	"  `maxKeys` bigint(20) unsigned DEFAULT '0' COMMENT '最多Key值',\n" +
	"  `maxSize` json DEFAULT NULL COMMENT '最大缓存内容尺寸',\n" +
	"  `type` varchar(255) DEFAULT NULL COMMENT '存储类型',\n" +
	"  `options` json DEFAULT NULL COMMENT '存储选项',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HTTP缓存策略';\n" +
	"CREATE TABLE `edgeHTTPFirewallPolicies` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  `inbound` json DEFAULT NULL COMMENT '入站规则',\n" +
	"  `outbound` json DEFAULT NULL COMMENT '出站规则',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HTTP防火墙';\n" +
	"CREATE TABLE `edgeHTTPFirewallRuleGroups` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  `code` varchar(255) DEFAULT NULL COMMENT '代号',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `sets` json DEFAULT NULL COMMENT '规则集列表',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='防火墙规则分组';\n" +
	"CREATE TABLE `edgeHTTPFirewallRuleSets` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `code` varchar(255) DEFAULT NULL COMMENT '代号',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `rules` json DEFAULT NULL COMMENT '规则列表',\n" +
	"  `connector` varchar(64) DEFAULT NULL COMMENT '规则之间的关系',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `action` varchar(255) DEFAULT NULL COMMENT '执行的动作',\n" +
	"  `actionOptions` json DEFAULT NULL COMMENT '动作的选项',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='防火墙规则集';\n" +
	"CREATE TABLE `edgeHTTPFirewallRules` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '说明',\n" +
	"  `param` varchar(1024) DEFAULT NULL COMMENT '参数',\n" +
	"  `operator` varchar(255) DEFAULT NULL COMMENT '操作符',\n" +
	"  `value` varchar(1024) DEFAULT NULL COMMENT '对比值',\n" +
	"  `isCaseInsensitive` tinyint(1) unsigned DEFAULT '1' COMMENT '是否大小写不敏感',\n" +
	"  `checkpointOptions` json DEFAULT NULL COMMENT '检查点参数',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='防火墙规则';\n" +
	"CREATE TABLE `edgeHTTPGzips` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `level` int(11) unsigned DEFAULT '0' COMMENT '压缩级别',\n" +
	"  `minLength` json DEFAULT NULL COMMENT '可压缩最小值',\n" +
	"  `maxLength` json DEFAULT NULL COMMENT '可压缩最大值',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `conds` json DEFAULT NULL COMMENT '条件',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Gzip配置';\n" +
	"CREATE TABLE `edgeHTTPHeaderPolicies` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `isOn` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '是否启用',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `addHeaders` json DEFAULT NULL COMMENT '添加的Header',\n" +
	"  `addTrailers` json DEFAULT NULL COMMENT '添加的Trailers',\n" +
	"  `setHeaders` json DEFAULT NULL COMMENT '设置Header',\n" +
	"  `replaceHeaders` json DEFAULT NULL COMMENT '替换Header内容',\n" +
	"  `expires` json DEFAULT NULL COMMENT 'Expires单独设置',\n" +
	"  `deleteHeaders` json DEFAULT NULL COMMENT '删除的Headers',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Header定义';\n" +
	"CREATE TABLE `edgeHTTPHeaders` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `value` varchar(1024) DEFAULT NULL COMMENT '值',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  `status` json DEFAULT NULL COMMENT '状态码设置',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HTTP Header';\n" +
	"CREATE TABLE `edgeHTTPLocations` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `parentId` int(11) unsigned DEFAULT '0' COMMENT '父级ID',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `pattern` varchar(1024) DEFAULT NULL COMMENT '匹配规则',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  `webId` int(11) unsigned DEFAULT '0' COMMENT 'Web配置ID',\n" +
	"  `reverseProxy` json DEFAULT NULL COMMENT '反向代理',\n" +
	"  `urlPrefix` varchar(1024) DEFAULT NULL COMMENT 'URL前缀',\n" +
	"  `isBreak` tinyint(1) unsigned DEFAULT '0' COMMENT '是否终止匹配',\n" +
	"  `conds` json DEFAULT NULL COMMENT '匹配条件',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路径规则配置';\n" +
	"CREATE TABLE `edgeHTTPPages` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '0' COMMENT '是否启用',\n" +
	"  `statusList` json DEFAULT NULL COMMENT '状态列表',\n" +
	"  `url` varchar(1024) DEFAULT NULL COMMENT '页面URL',\n" +
	"  `newStatus` int(3) DEFAULT NULL COMMENT '新状态码',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='特殊页面';\n" +
	"CREATE TABLE `edgeHTTPRewriteRules` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `pattern` varchar(1024) DEFAULT NULL COMMENT '匹配规则',\n" +
	"  `replace` varchar(1024) DEFAULT NULL COMMENT '跳转后的地址',\n" +
	"  `mode` varchar(255) DEFAULT NULL COMMENT '替换模式',\n" +
	"  `redirectStatus` int(3) unsigned DEFAULT '0' COMMENT '跳转的状态码',\n" +
	"  `proxyHost` varchar(255) DEFAULT NULL COMMENT '代理的主机名',\n" +
	"  `isBreak` tinyint(1) unsigned DEFAULT '1' COMMENT '是否终止解析',\n" +
	"  `withQuery` tinyint(1) unsigned DEFAULT '1' COMMENT '是否保留URI参数',\n" +
	"  `conds` json DEFAULT NULL COMMENT '匹配条件',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='重写规则';\n" +
	"CREATE TABLE `edgeHTTPWebs` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `root` json DEFAULT NULL COMMENT '根目录',\n" +
	"  `charset` json DEFAULT NULL COMMENT '字符集',\n" +
	"  `shutdown` json DEFAULT NULL COMMENT '临时关闭页面配置',\n" +
	"  `pages` json DEFAULT NULL COMMENT '特殊页面',\n" +
	"  `redirectToHttps` json DEFAULT NULL COMMENT '跳转到HTTPS设置',\n" +
	"  `indexes` json DEFAULT NULL COMMENT '首页文件列表',\n" +
	"  `maxRequestBodySize` json DEFAULT NULL COMMENT '最大允许的请求内容尺寸',\n" +
	"  `requestHeader` json DEFAULT NULL COMMENT '请求Header配置',\n" +
	"  `responseHeader` json DEFAULT NULL COMMENT '响应Header配置',\n" +
	"  `accessLog` json DEFAULT NULL COMMENT '访问日志配置',\n" +
	"  `stat` json DEFAULT NULL COMMENT '统计配置',\n" +
	"  `gzip` json DEFAULT NULL COMMENT 'Gzip配置',\n" +
	"  `cache` json DEFAULT NULL COMMENT '缓存配置',\n" +
	"  `firewall` json DEFAULT NULL COMMENT '防火墙设置',\n" +
	"  `locations` json DEFAULT NULL COMMENT '路径规则配置',\n" +
	"  `websocket` json DEFAULT NULL COMMENT 'Websocket设置',\n" +
	"  `rewriteRules` json DEFAULT NULL COMMENT '重写规则配置',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HTTP Web';\n" +
	"CREATE TABLE `edgeHTTPWebsockets` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `handshakeTimeout` json DEFAULT NULL COMMENT '握手超时时间',\n" +
	"  `allowAllOrigins` tinyint(1) unsigned DEFAULT '1' COMMENT '是否支持所有源',\n" +
	"  `allowedOrigins` json DEFAULT NULL COMMENT '支持的源域名列表',\n" +
	"  `requestSameOrigin` tinyint(1) unsigned DEFAULT '1' COMMENT '是否请求一样的Origin',\n" +
	"  `requestOrigin` varchar(255) DEFAULT NULL COMMENT '请求Origin',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Websocket设置';\n" +
	"CREATE TABLE `edgeLogs` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `level` varchar(32) DEFAULT NULL COMMENT '级别',\n" +
	"  `description` varchar(255) DEFAULT NULL COMMENT '描述',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `action` varchar(255) DEFAULT NULL COMMENT '动作',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `providerId` int(11) unsigned DEFAULT '0' COMMENT '供应商ID',\n" +
	"  `ip` varchar(32) DEFAULT NULL COMMENT 'IP地址',\n" +
	"  `type` varchar(255) DEFAULT 'admin' COMMENT '类型：admin, user',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志';\n" +
	"CREATE TABLE `edgeNodeClusters` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `useAllAPINodes` tinyint(1) unsigned DEFAULT '1' COMMENT '是否使用所有API节点',\n" +
	"  `apiNodes` json DEFAULT NULL COMMENT '使用的API节点',\n" +
	"  `installDir` varchar(512) DEFAULT NULL COMMENT '安装目录',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `grantId` int(11) unsigned DEFAULT '0' COMMENT '默认认证方式',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点集群';\n" +
	"CREATE TABLE `edgeNodeGrants` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `method` varchar(64) DEFAULT NULL COMMENT '登录方式',\n" +
	"  `username` varchar(255) DEFAULT NULL COMMENT '用户名',\n" +
	"  `password` varchar(255) DEFAULT NULL COMMENT '密码',\n" +
	"  `su` tinyint(1) unsigned DEFAULT '1' COMMENT '是否需要su',\n" +
	"  `privateKey` varchar(4096) DEFAULT NULL COMMENT '密钥',\n" +
	"  `description` varchar(255) DEFAULT NULL COMMENT '备注',\n" +
	"  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '专有节点',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点授权';\n" +
	"CREATE TABLE `edgeNodeGroups` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点分组';\n" +
	"CREATE TABLE `edgeNodeIPAddresses` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `ip` varchar(128) DEFAULT NULL COMMENT 'IP地址',\n" +
	"  `description` varchar(255) DEFAULT NULL COMMENT '描述',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `nodeId` (`nodeId`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点IP地址';\n" +
	"CREATE TABLE `edgeNodeLogins` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `type` varchar(255) DEFAULT NULL COMMENT '类型：ssh,agent',\n" +
	"  `params` json DEFAULT NULL COMMENT '配置参数',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `nodeId` (`nodeId`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点登录信息';\n" +
	"CREATE TABLE `edgeNodeLogs` (\n" +
	"  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `role` varchar(64) DEFAULT NULL COMMENT '节点角色',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `tag` varchar(255) DEFAULT NULL COMMENT '标签',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  `level` varchar(32) DEFAULT NULL COMMENT '级别',\n" +
	"  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',\n" +
	"  `day` varchar(8) DEFAULT NULL COMMENT '日期',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `level` (`level`),\n" +
	"  KEY `day` (`day`),\n" +
	"  KEY `role_nodeId` (`role`,`nodeId`) USING BTREE\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点日志';\n" +
	"CREATE TABLE `edgeNodeRegions` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点区域';\n" +
	"CREATE TABLE `edgeNodes` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `uniqueId` varchar(32) DEFAULT NULL COMMENT '节点ID',\n" +
	"  `secret` varchar(32) DEFAULT NULL COMMENT '密钥',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '节点名',\n" +
	"  `code` varchar(255) DEFAULT NULL COMMENT '代号',\n" +
	"  `clusterId` int(11) unsigned DEFAULT '0' COMMENT '集群ID',\n" +
	"  `regionId` int(11) unsigned DEFAULT '0' COMMENT '区域ID',\n" +
	"  `groupId` int(11) unsigned DEFAULT '0' COMMENT '分组ID',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `status` json DEFAULT NULL COMMENT '最新的状态',\n" +
	"  `version` int(11) unsigned DEFAULT '0' COMMENT '当前版本号',\n" +
	"  `latestVersion` int(11) unsigned DEFAULT '0' COMMENT '最后版本号',\n" +
	"  `installDir` varchar(512) DEFAULT NULL COMMENT '安装目录',\n" +
	"  `isInstalled` tinyint(1) unsigned DEFAULT '0' COMMENT '是否已安装',\n" +
	"  `installStatus` json DEFAULT NULL COMMENT '安装状态',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `connectedAPINodes` json DEFAULT NULL COMMENT '当前连接的API节点',\n" +
	"  `maxCPU` int(4) unsigned DEFAULT '0' COMMENT '可以使用的最多CPU',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `uniqueId` (`uniqueId`),\n" +
	"  KEY `clusterId` (`clusterId`),\n" +
	"  KEY `groupId` (`groupId`),\n" +
	"  KEY `regionId` (`regionId`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点';\n" +
	"CREATE TABLE `edgeOrigins` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `version` int(11) unsigned DEFAULT '0' COMMENT '版本',\n" +
	"  `addr` json DEFAULT NULL COMMENT '地址',\n" +
	"  `description` varchar(512) DEFAULT NULL COMMENT '描述',\n" +
	"  `code` varchar(255) DEFAULT NULL COMMENT '代号',\n" +
	"  `weight` int(11) unsigned DEFAULT '0' COMMENT '权重',\n" +
	"  `connTimeout` json DEFAULT NULL COMMENT '连接超时',\n" +
	"  `readTimeout` json DEFAULT NULL COMMENT '读超时',\n" +
	"  `idleTimeout` json DEFAULT NULL COMMENT '空闲连接超时',\n" +
	"  `maxFails` int(11) unsigned DEFAULT '0' COMMENT '最多失败次数',\n" +
	"  `maxConns` int(11) unsigned DEFAULT '0' COMMENT '最大并发连接数',\n" +
	"  `maxIdleConns` int(11) unsigned DEFAULT '0' COMMENT '最多空闲连接数',\n" +
	"  `httpRequestURI` varchar(1024) DEFAULT NULL COMMENT '转发后的请求URI',\n" +
	"  `httpRequestHeader` json DEFAULT NULL COMMENT '请求Header配置',\n" +
	"  `httpResponseHeader` json DEFAULT NULL COMMENT '响应Header配置',\n" +
	"  `host` varchar(255) DEFAULT NULL COMMENT '自定义主机名',\n" +
	"  `healthCheck` json DEFAULT NULL COMMENT '健康检查设置',\n" +
	"  `cert` json DEFAULT NULL COMMENT '证书设置',\n" +
	"  `ftp` json DEFAULT NULL COMMENT 'FTP相关设置',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='源站';\n" +
	"CREATE TABLE `edgeProviders` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `username` varchar(64) DEFAULT NULL COMMENT '用户名',\n" +
	"  `password` varchar(32) DEFAULT NULL COMMENT '密码',\n" +
	"  `fullname` varchar(64) DEFAULT NULL COMMENT '真实姓名',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `updatedAt` bigint(11) unsigned DEFAULT '0' COMMENT '修改时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商';\n" +
	"CREATE TABLE `edgeReverseProxies` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `scheduling` json DEFAULT NULL COMMENT '调度算法',\n" +
	"  `primaryOrigins` json DEFAULT NULL COMMENT '主要源站',\n" +
	"  `backupOrigins` json DEFAULT NULL COMMENT '备用源站',\n" +
	"  `stripPrefix` varchar(255) DEFAULT NULL COMMENT '去除URL前缀',\n" +
	"  `requestHost` varchar(255) DEFAULT NULL COMMENT '请求Host',\n" +
	"  `requestURI` varchar(1024) DEFAULT NULL COMMENT '请求URI',\n" +
	"  `autoFlush` tinyint(1) unsigned DEFAULT '0' COMMENT '是否自动刷新缓冲区',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='反向代理配置';\n" +
	"CREATE TABLE `edgeSSLCertGroups` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '分组名',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '分组排序',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '0' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='证书分组';\n" +
	"CREATE TABLE `edgeSSLCerts` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `updatedAt` bigint(11) unsigned DEFAULT '0' COMMENT '修改时间',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '证书名',\n" +
	"  `description` varchar(1024) DEFAULT NULL COMMENT '描述',\n" +
	"  `certData` blob COMMENT '证书内容',\n" +
	"  `keyData` blob COMMENT '密钥内容',\n" +
	"  `serverName` varchar(255) DEFAULT NULL COMMENT '证书使用的主机名',\n" +
	"  `isCA` tinyint(1) unsigned DEFAULT '0' COMMENT '是否为CA证书',\n" +
	"  `groupIds` json DEFAULT NULL COMMENT '证书分组',\n" +
	"  `timeBeginAt` bigint(11) unsigned DEFAULT '0' COMMENT '开始时间',\n" +
	"  `timeEndAt` bigint(11) unsigned DEFAULT '0' COMMENT '结束时间',\n" +
	"  `dnsNames` json DEFAULT NULL COMMENT 'DNS名称列表',\n" +
	"  `commonNames` json DEFAULT NULL COMMENT '发行单位列表',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SSL证书';\n" +
	"CREATE TABLE `edgeSSLPolicies` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `certs` json DEFAULT NULL COMMENT '证书列表',\n" +
	"  `clientCACerts` json DEFAULT NULL COMMENT '客户端证书',\n" +
	"  `clientAuthType` int(11) unsigned DEFAULT '0' COMMENT '客户端认证类型',\n" +
	"  `minVersion` varchar(32) DEFAULT NULL COMMENT '支持的SSL最小版本',\n" +
	"  `cipherSuitesIsOn` tinyint(1) unsigned DEFAULT '0' COMMENT '是否自定义加密算法套件',\n" +
	"  `cipherSuites` json DEFAULT NULL COMMENT '加密算法套件',\n" +
	"  `hsts` json DEFAULT NULL COMMENT 'HSTS设置',\n" +
	"  `http2Enabled` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用HTTP/2',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SSL配置策略';\n" +
	"CREATE TABLE `edgeServerGroups` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务分组';\n" +
	"CREATE TABLE `edgeServers` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',\n" +
	"  `type` varchar(64) DEFAULT NULL COMMENT '服务类型',\n" +
	"  `name` varchar(255) DEFAULT NULL COMMENT '名称',\n" +
	"  `description` varchar(512) DEFAULT NULL COMMENT '描述',\n" +
	"  `serverNames` json DEFAULT NULL COMMENT '域名列表',\n" +
	"  `http` json DEFAULT NULL COMMENT 'HTTP配置',\n" +
	"  `https` json DEFAULT NULL COMMENT 'HTTPS配置',\n" +
	"  `tcp` json DEFAULT NULL COMMENT 'TCP配置',\n" +
	"  `tls` json DEFAULT NULL COMMENT 'TLS配置',\n" +
	"  `unix` json DEFAULT NULL COMMENT 'Unix配置',\n" +
	"  `udp` json DEFAULT NULL COMMENT 'UDP配置',\n" +
	"  `webId` int(11) unsigned DEFAULT '0' COMMENT 'WEB配置',\n" +
	"  `reverseProxy` json DEFAULT NULL COMMENT '反向代理配置',\n" +
	"  `groupIds` json DEFAULT NULL COMMENT '分组ID列表',\n" +
	"  `config` json DEFAULT NULL COMMENT '服务配置，自动生成',\n" +
	"  `configMd5` varchar(32) DEFAULT NULL COMMENT 'Md5',\n" +
	"  `clusterId` int(11) unsigned DEFAULT '0' COMMENT '集群ID',\n" +
	"  `includeNodes` json DEFAULT NULL COMMENT '部署条件',\n" +
	"  `excludeNodes` json DEFAULT NULL COMMENT '节点排除条件',\n" +
	"  `version` int(11) unsigned DEFAULT '0' COMMENT '版本号',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  KEY `userId` (`userId`),\n" +
	"  KEY `adminId` (`adminId`),\n" +
	"  KEY `isUpdating_state` (`state`) USING BTREE\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务';\n" +
	"CREATE TABLE `edgeSysEvents` (\n" +
	"  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `type` varchar(255) DEFAULT NULL COMMENT '类型',\n" +
	"  `params` json DEFAULT NULL COMMENT '参数',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统事件';\n" +
	"CREATE TABLE `edgeSysLockers` (\n" +
	"  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `key` varchar(255) DEFAULT NULL COMMENT '键值',\n" +
	"  `version` bigint(20) unsigned DEFAULT '0' COMMENT '版本号',\n" +
	"  `timeoutAt` bigint(11) unsigned DEFAULT '0' COMMENT '超时时间',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  UNIQUE KEY `key` (`key`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='并发锁';\n" +
	"CREATE TABLE `edgeSysSettings` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `code` varchar(255) DEFAULT NULL COMMENT '代号',\n" +
	"  `value` json DEFAULT NULL COMMENT '配置值',\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  UNIQUE KEY `code` (`code`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置';\n" +
	"CREATE TABLE `edgeTCPFirewallPolicies` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `adminId` int(11) DEFAULT NULL COMMENT '管理员ID',\n" +
	"  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',\n" +
	"  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='TCP防火墙';\n" +
	"CREATE TABLE `edgeUsers` (\n" +
	"  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `username` varchar(64) DEFAULT NULL COMMENT '用户名',\n" +
	"  `password` varchar(32) DEFAULT NULL COMMENT '密码',\n" +
	"  `fullname` varchar(64) DEFAULT NULL COMMENT '真实姓名',\n" +
	"  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',\n" +
	"  `updatedAt` bigint(11) unsigned DEFAULT '0' COMMENT '修改时间',\n" +
	"  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户';\n" +
	"CREATE TABLE `edgeVersions` (\n" +
	"  `id` bigint(16) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',\n" +
	"  `version` varchar(64) DEFAULT NULL,\n" +
	"  PRIMARY KEY (`id`)\n" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库结构版本';\n" +
	"\n"
