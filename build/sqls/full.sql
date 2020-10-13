CREATE TABLE `edgeAPINodes` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `clusterId` int(11) unsigned DEFAULT '0' COMMENT '专用集群ID',
  `uniqueId` varchar(32) DEFAULT NULL COMMENT '唯一ID',
  `secret` varchar(32) DEFAULT NULL COMMENT '密钥',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `http` json DEFAULT NULL COMMENT '监听的HTTP配置',
  `https` json DEFAULT NULL COMMENT '监听的HTTPS配置',
  `accessAddrs` json DEFAULT NULL COMMENT '外部访问地址',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `weight` int(11) unsigned DEFAULT '0' COMMENT '权重',
  `status` json DEFAULT NULL COMMENT '运行状态',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniqueId` (`uniqueId`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API节点';
CREATE TABLE `edgeAPITokens` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `nodeId` varchar(32) DEFAULT NULL COMMENT '节点ID',
  `secret` varchar(255) DEFAULT NULL COMMENT '节点密钥',
  `role` varchar(64) DEFAULT NULL COMMENT '节点角色',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`),
  KEY `nodeId` (`nodeId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API令牌管理';
CREATE TABLE `edgeAdmins` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `username` varchar(64) DEFAULT NULL COMMENT '用户名',
  `password` varchar(32) DEFAULT NULL COMMENT '密码',
  `fullname` varchar(64) DEFAULT NULL COMMENT '全名',
  `isSuper` tinyint(1) unsigned DEFAULT '0' COMMENT '是否为超级管理员',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `updatedAt` bigint(11) unsigned DEFAULT '0' COMMENT '修改时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='管理员';
CREATE TABLE `edgeDBNodes` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `role` varchar(255) DEFAULT NULL COMMENT '数据库角色',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `host` varchar(255) DEFAULT NULL COMMENT '主机',
  `port` int(11) unsigned DEFAULT '0' COMMENT '端口',
  `database` varchar(255) DEFAULT NULL COMMENT '数据库名称',
  `username` varchar(255) DEFAULT NULL COMMENT '用户名',
  `password` varchar(255) DEFAULT NULL COMMENT '密码',
  `charset` varchar(255) DEFAULT NULL COMMENT '通讯字符集',
  `connTimeout` int(11) unsigned DEFAULT '0' COMMENT '连接超时时间（秒）',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `weight` int(11) unsigned DEFAULT '0' COMMENT '权重',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库节点';
CREATE TABLE `edgeFileChunks` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `fileId` int(11) unsigned DEFAULT '0' COMMENT '文件ID',
  `data` longblob COMMENT '分块内容',
  PRIMARY KEY (`id`),
  KEY `fileId` (`fileId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件片段';
CREATE TABLE `edgeFiles` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `description` varchar(255) DEFAULT NULL COMMENT '文件描述',
  `filename` varchar(255) DEFAULT NULL COMMENT '文件名',
  `size` int(11) unsigned DEFAULT '0' COMMENT '文件尺寸',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  `type` varchar(64) DEFAULT '' COMMENT '类型',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`),
  KEY `type` (`type`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文件管理';
CREATE TABLE `edgeHTTPAccessLogPolicies` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `type` varchar(255) DEFAULT NULL COMMENT '存储类型',
  `options` json DEFAULT NULL COMMENT '存储选项',
  `conds` json DEFAULT NULL COMMENT '请求条件',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='访问日志策略';
CREATE TABLE `edgeHTTPAccessLogs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `serverId` int(11) unsigned DEFAULT '0' COMMENT '服务ID',
  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',
  `status` int(3) unsigned DEFAULT '0' COMMENT '状态码',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `content` json DEFAULT NULL COMMENT '日志内容',
  `requestId` varchar(128) DEFAULT NULL COMMENT '请求ID',
  PRIMARY KEY (`id`),
  KEY `serverId` (`serverId`),
  KEY `nodeId` (`nodeId`),
  KEY `serverId_status` (`serverId`,`status`),
  KEY `requestId` (`requestId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE `edgeHTTPAccessLogs_20201010` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `serverId` int(11) unsigned DEFAULT '0' COMMENT '服务ID',
  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',
  `status` int(3) unsigned DEFAULT '0' COMMENT '状态码',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `content` json DEFAULT NULL COMMENT '日志内容',
  `day` varchar(8) DEFAULT NULL COMMENT '日期Ymd',
  PRIMARY KEY (`id`),
  KEY `serverId` (`serverId`),
  KEY `nodeId` (`nodeId`),
  KEY `serverId_status` (`serverId`,`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
CREATE TABLE `edgeHTTPCachePolicies` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `capacity` json DEFAULT NULL COMMENT '容量数据',
  `maxKeys` bigint(20) unsigned DEFAULT '0' COMMENT '最多Key值',
  `maxSize` json DEFAULT NULL COMMENT '最大缓存内容尺寸',
  `type` varchar(255) DEFAULT NULL COMMENT '存储类型',
  `options` json DEFAULT NULL COMMENT '存储选项',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HTTP缓存策略';
CREATE TABLE `edgeHTTPFirewallPolicies` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `inbound` json DEFAULT NULL COMMENT '入站规则',
  `outbound` json DEFAULT NULL COMMENT '出站规则',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HTTP防火墙';
CREATE TABLE `edgeHTTPFirewallRuleGroups` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `code` varchar(255) DEFAULT NULL COMMENT '代号',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `sets` json DEFAULT NULL COMMENT '规则集列表',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='防火墙规则分组';
CREATE TABLE `edgeHTTPFirewallRuleSets` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `code` varchar(255) DEFAULT NULL COMMENT '代号',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `rules` json DEFAULT NULL COMMENT '规则列表',
  `connector` varchar(64) DEFAULT NULL COMMENT '规则之间的关系',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `action` varchar(255) DEFAULT NULL COMMENT '执行的动作',
  `actionOptions` json DEFAULT NULL COMMENT '动作的选项',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='防火墙规则集';
CREATE TABLE `edgeHTTPFirewallRules` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `description` varchar(1024) DEFAULT NULL COMMENT '说明',
  `param` varchar(1024) DEFAULT NULL COMMENT '参数',
  `operator` varchar(255) DEFAULT NULL COMMENT '操作符',
  `value` varchar(1024) DEFAULT NULL COMMENT '对比值',
  `isCaseInsensitive` tinyint(1) unsigned DEFAULT '1' COMMENT '是否大小写不敏感',
  `checkpointOptions` json DEFAULT NULL COMMENT '检查点参数',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='防火墙规则';
CREATE TABLE `edgeHTTPGzips` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `level` int(11) unsigned DEFAULT '0' COMMENT '压缩级别',
  `minLength` json DEFAULT NULL COMMENT '可压缩最小值',
  `maxLength` json DEFAULT NULL COMMENT '可压缩最大值',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `conds` json DEFAULT NULL COMMENT '条件',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Gzip配置';
CREATE TABLE `edgeHTTPHeaderPolicies` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `isOn` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '是否启用',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `addHeaders` json DEFAULT NULL COMMENT '添加的Header',
  `addTrailers` json DEFAULT NULL COMMENT '添加的Trailers',
  `setHeaders` json DEFAULT NULL COMMENT '设置Header',
  `replaceHeaders` json DEFAULT NULL COMMENT '替换Header内容',
  `expires` json DEFAULT NULL COMMENT 'Expires单独设置',
  `deleteHeaders` json DEFAULT NULL COMMENT '删除的Headers',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Header定义';
CREATE TABLE `edgeHTTPHeaders` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `value` varchar(1024) DEFAULT NULL COMMENT '值',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  `status` json DEFAULT NULL COMMENT '状态码设置',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HTTP Header';
CREATE TABLE `edgeHTTPLocations` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `parentId` int(11) unsigned DEFAULT '0' COMMENT '父级ID',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `pattern` varchar(1024) DEFAULT NULL COMMENT '匹配规则',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `webId` int(11) unsigned DEFAULT '0' COMMENT 'Web配置ID',
  `reverseProxy` json DEFAULT NULL COMMENT '反向代理',
  `urlPrefix` varchar(1024) DEFAULT NULL COMMENT 'URL前缀',
  `isBreak` tinyint(1) unsigned DEFAULT '0' COMMENT '是否终止匹配',
  `conds` json DEFAULT NULL COMMENT '匹配条件',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='路径规则配置';
CREATE TABLE `edgeHTTPPages` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `isOn` tinyint(1) unsigned DEFAULT '0' COMMENT '是否启用',
  `statusList` json DEFAULT NULL COMMENT '状态列表',
  `url` varchar(1024) DEFAULT NULL COMMENT '页面URL',
  `newStatus` int(3) DEFAULT NULL COMMENT '新状态码',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='特殊页面';
CREATE TABLE `edgeHTTPRewriteRules` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `pattern` varchar(1024) DEFAULT NULL COMMENT '匹配规则',
  `replace` varchar(1024) DEFAULT NULL COMMENT '跳转后的地址',
  `mode` varchar(255) DEFAULT NULL COMMENT '替换模式',
  `redirectStatus` int(3) unsigned DEFAULT '0' COMMENT '跳转的状态码',
  `proxyHost` varchar(255) DEFAULT NULL COMMENT '代理的主机名',
  `isBreak` tinyint(1) unsigned DEFAULT '1' COMMENT '是否终止解析',
  `withQuery` tinyint(1) unsigned DEFAULT '1' COMMENT '是否保留URI参数',
  `conds` json DEFAULT NULL COMMENT '匹配条件',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='重写规则';
CREATE TABLE `edgeHTTPWebs` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `root` json DEFAULT NULL COMMENT '根目录',
  `charset` json DEFAULT NULL COMMENT '字符集',
  `shutdown` json DEFAULT NULL COMMENT '临时关闭页面配置',
  `pages` json DEFAULT NULL COMMENT '特殊页面',
  `redirectToHttps` json DEFAULT NULL COMMENT '跳转到HTTPS设置',
  `indexes` json DEFAULT NULL COMMENT '首页文件列表',
  `maxRequestBodySize` json DEFAULT NULL COMMENT '最大允许的请求内容尺寸',
  `requestHeader` json DEFAULT NULL COMMENT '请求Header配置',
  `responseHeader` json DEFAULT NULL COMMENT '响应Header配置',
  `accessLog` json DEFAULT NULL COMMENT '访问日志配置',
  `stat` json DEFAULT NULL COMMENT '统计配置',
  `gzip` json DEFAULT NULL COMMENT 'Gzip配置',
  `cache` json DEFAULT NULL COMMENT '缓存配置',
  `firewall` json DEFAULT NULL COMMENT '防火墙设置',
  `locations` json DEFAULT NULL COMMENT '路径规则配置',
  `websocket` json DEFAULT NULL COMMENT 'Websocket设置',
  `rewriteRules` json DEFAULT NULL COMMENT '重写规则配置',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='HTTP Web';
CREATE TABLE `edgeHTTPWebsockets` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `handshakeTimeout` json DEFAULT NULL COMMENT '握手超时时间',
  `allowAllOrigins` tinyint(1) unsigned DEFAULT '1' COMMENT '是否支持所有源',
  `allowedOrigins` json DEFAULT NULL COMMENT '支持的源域名列表',
  `requestSameOrigin` tinyint(1) unsigned DEFAULT '1' COMMENT '是否请求一样的Origin',
  `requestOrigin` varchar(255) DEFAULT NULL COMMENT '请求Origin',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Websocket设置';
CREATE TABLE `edgeLogs` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `level` varchar(32) DEFAULT NULL COMMENT '级别',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `action` varchar(255) DEFAULT NULL COMMENT '动作',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `providerId` int(11) unsigned DEFAULT '0' COMMENT '供应商ID',
  `ip` varchar(32) DEFAULT NULL COMMENT 'IP地址',
  `type` varchar(255) DEFAULT 'admin' COMMENT '类型：admin, user',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='操作日志';
CREATE TABLE `edgeNodeClusters` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `useAllAPINodes` tinyint(1) unsigned DEFAULT '1' COMMENT '是否使用所有API节点',
  `apiNodes` json DEFAULT NULL COMMENT '使用的API节点',
  `installDir` varchar(512) DEFAULT NULL COMMENT '安装目录',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `grantId` int(11) unsigned DEFAULT '0' COMMENT '默认认证方式',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点集群';
CREATE TABLE `edgeNodeGrants` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `method` varchar(64) DEFAULT NULL COMMENT '登录方式',
  `username` varchar(255) DEFAULT NULL COMMENT '用户名',
  `password` varchar(255) DEFAULT NULL COMMENT '密码',
  `su` tinyint(1) unsigned DEFAULT '1' COMMENT '是否需要su',
  `privateKey` varchar(4096) DEFAULT NULL COMMENT '密钥',
  `description` varchar(255) DEFAULT NULL COMMENT '备注',
  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '专有节点',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点授权';
CREATE TABLE `edgeNodeGroups` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点分组';
CREATE TABLE `edgeNodeIPAddresses` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `ip` varchar(128) DEFAULT NULL COMMENT 'IP地址',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  PRIMARY KEY (`id`),
  KEY `nodeId` (`nodeId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点IP地址';
CREATE TABLE `edgeNodeLogins` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `type` varchar(255) DEFAULT NULL COMMENT '类型：ssh,agent',
  `params` json DEFAULT NULL COMMENT '配置参数',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`),
  KEY `nodeId` (`nodeId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点登录信息';
CREATE TABLE `edgeNodeLogs` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `role` varchar(64) DEFAULT NULL COMMENT '节点角色',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `tag` varchar(255) DEFAULT NULL COMMENT '标签',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `level` varchar(32) DEFAULT NULL COMMENT '级别',
  `nodeId` int(11) unsigned DEFAULT '0' COMMENT '节点ID',
  `day` varchar(8) DEFAULT NULL COMMENT '日期',
  PRIMARY KEY (`id`),
  KEY `level` (`level`),
  KEY `day` (`day`),
  KEY `role_nodeId` (`role`,`nodeId`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点日志';
CREATE TABLE `edgeNodeRegions` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点区域';
CREATE TABLE `edgeNodes` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `uniqueId` varchar(32) DEFAULT NULL COMMENT '节点ID',
  `secret` varchar(32) DEFAULT NULL COMMENT '密钥',
  `name` varchar(255) DEFAULT NULL COMMENT '节点名',
  `code` varchar(255) DEFAULT NULL COMMENT '代号',
  `clusterId` int(11) unsigned DEFAULT '0' COMMENT '集群ID',
  `regionId` int(11) unsigned DEFAULT '0' COMMENT '区域ID',
  `groupId` int(11) unsigned DEFAULT '0' COMMENT '分组ID',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `status` json DEFAULT NULL COMMENT '最新的状态',
  `version` int(11) unsigned DEFAULT '0' COMMENT '当前版本号',
  `latestVersion` int(11) unsigned DEFAULT '0' COMMENT '最后版本号',
  `installDir` varchar(512) DEFAULT NULL COMMENT '安装目录',
  `isInstalled` tinyint(1) unsigned DEFAULT '0' COMMENT '是否已安装',
  `installStatus` json DEFAULT NULL COMMENT '安装状态',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `connectedAPINodes` json DEFAULT NULL COMMENT '当前连接的API节点',
  `maxCPU` int(4) unsigned DEFAULT '0' COMMENT '可以使用的最多CPU',
  PRIMARY KEY (`id`),
  KEY `uniqueId` (`uniqueId`),
  KEY `clusterId` (`clusterId`),
  KEY `groupId` (`groupId`),
  KEY `regionId` (`regionId`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='节点';
CREATE TABLE `edgeOrigins` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `version` int(11) unsigned DEFAULT '0' COMMENT '版本',
  `addr` json DEFAULT NULL COMMENT '地址',
  `description` varchar(512) DEFAULT NULL COMMENT '描述',
  `code` varchar(255) DEFAULT NULL COMMENT '代号',
  `weight` int(11) unsigned DEFAULT '0' COMMENT '权重',
  `connTimeout` json DEFAULT NULL COMMENT '连接超时',
  `readTimeout` json DEFAULT NULL COMMENT '读超时',
  `idleTimeout` json DEFAULT NULL COMMENT '空闲连接超时',
  `maxFails` int(11) unsigned DEFAULT '0' COMMENT '最多失败次数',
  `maxConns` int(11) unsigned DEFAULT '0' COMMENT '最大并发连接数',
  `maxIdleConns` int(11) unsigned DEFAULT '0' COMMENT '最多空闲连接数',
  `httpRequestURI` varchar(1024) DEFAULT NULL COMMENT '转发后的请求URI',
  `httpRequestHeader` json DEFAULT NULL COMMENT '请求Header配置',
  `httpResponseHeader` json DEFAULT NULL COMMENT '响应Header配置',
  `host` varchar(255) DEFAULT NULL COMMENT '自定义主机名',
  `healthCheck` json DEFAULT NULL COMMENT '健康检查设置',
  `cert` json DEFAULT NULL COMMENT '证书设置',
  `ftp` json DEFAULT NULL COMMENT 'FTP相关设置',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='源站';
CREATE TABLE `edgeProviders` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `username` varchar(64) DEFAULT NULL COMMENT '用户名',
  `password` varchar(32) DEFAULT NULL COMMENT '密码',
  `fullname` varchar(64) DEFAULT NULL COMMENT '真实姓名',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `updatedAt` bigint(11) unsigned DEFAULT '0' COMMENT '修改时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='供应商';
CREATE TABLE `edgeReverseProxies` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `scheduling` json DEFAULT NULL COMMENT '调度算法',
  `primaryOrigins` json DEFAULT NULL COMMENT '主要源站',
  `backupOrigins` json DEFAULT NULL COMMENT '备用源站',
  `stripPrefix` varchar(255) DEFAULT NULL COMMENT '去除URL前缀',
  `requestHost` varchar(255) DEFAULT NULL COMMENT '请求Host',
  `requestURI` varchar(1024) DEFAULT NULL COMMENT '请求URI',
  `autoFlush` tinyint(1) unsigned DEFAULT '0' COMMENT '是否自动刷新缓冲区',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='反向代理配置';
CREATE TABLE `edgeSSLCertGroups` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `name` varchar(255) DEFAULT NULL COMMENT '分组名',
  `order` int(11) unsigned DEFAULT '0' COMMENT '分组排序',
  `state` tinyint(1) unsigned DEFAULT '0' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='证书分组';
CREATE TABLE `edgeSSLCerts` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `updatedAt` bigint(11) unsigned DEFAULT '0' COMMENT '修改时间',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `name` varchar(255) DEFAULT NULL COMMENT '证书名',
  `description` varchar(1024) DEFAULT NULL COMMENT '描述',
  `certData` blob COMMENT '证书内容',
  `keyData` blob COMMENT '密钥内容',
  `serverName` varchar(255) DEFAULT NULL COMMENT '证书使用的主机名',
  `isCA` tinyint(1) unsigned DEFAULT '0' COMMENT '是否为CA证书',
  `groupIds` json DEFAULT NULL COMMENT '证书分组',
  `timeBeginAt` bigint(11) unsigned DEFAULT '0' COMMENT '开始时间',
  `timeEndAt` bigint(11) unsigned DEFAULT '0' COMMENT '结束时间',
  `dnsNames` json DEFAULT NULL COMMENT 'DNS名称列表',
  `commonNames` json DEFAULT NULL COMMENT '发行单位列表',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SSL证书';
CREATE TABLE `edgeSSLPolicies` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `certs` json DEFAULT NULL COMMENT '证书列表',
  `clientCACerts` json DEFAULT NULL COMMENT '客户端证书',
  `clientAuthType` int(11) unsigned DEFAULT '0' COMMENT '客户端认证类型',
  `minVersion` varchar(32) DEFAULT NULL COMMENT '支持的SSL最小版本',
  `cipherSuitesIsOn` tinyint(1) unsigned DEFAULT '0' COMMENT '是否自定义加密算法套件',
  `cipherSuites` json DEFAULT NULL COMMENT '加密算法套件',
  `hsts` json DEFAULT NULL COMMENT 'HSTS设置',
  `http2Enabled` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用HTTP/2',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SSL配置策略';
CREATE TABLE `edgeServerGroups` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `order` int(11) unsigned DEFAULT '0' COMMENT '排序',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务分组';
CREATE TABLE `edgeServers` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `isOn` tinyint(1) unsigned DEFAULT '1' COMMENT '是否启用',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `adminId` int(11) unsigned DEFAULT '0' COMMENT '管理员ID',
  `type` varchar(64) DEFAULT NULL COMMENT '服务类型',
  `name` varchar(255) DEFAULT NULL COMMENT '名称',
  `description` varchar(512) DEFAULT NULL COMMENT '描述',
  `serverNames` json DEFAULT NULL COMMENT '域名列表',
  `http` json DEFAULT NULL COMMENT 'HTTP配置',
  `https` json DEFAULT NULL COMMENT 'HTTPS配置',
  `tcp` json DEFAULT NULL COMMENT 'TCP配置',
  `tls` json DEFAULT NULL COMMENT 'TLS配置',
  `unix` json DEFAULT NULL COMMENT 'Unix配置',
  `udp` json DEFAULT NULL COMMENT 'UDP配置',
  `webId` int(11) unsigned DEFAULT '0' COMMENT 'WEB配置',
  `reverseProxy` json DEFAULT NULL COMMENT '反向代理配置',
  `groupIds` json DEFAULT NULL COMMENT '分组ID列表',
  `config` json DEFAULT NULL COMMENT '服务配置，自动生成',
  `configMd5` varchar(32) DEFAULT NULL COMMENT 'Md5',
  `clusterId` int(11) unsigned DEFAULT '0' COMMENT '集群ID',
  `includeNodes` json DEFAULT NULL COMMENT '部署条件',
  `excludeNodes` json DEFAULT NULL COMMENT '节点排除条件',
  `version` int(11) unsigned DEFAULT '0' COMMENT '版本号',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`),
  KEY `userId` (`userId`),
  KEY `adminId` (`adminId`),
  KEY `isUpdating_state` (`state`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='服务';
CREATE TABLE `edgeSysEvents` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `type` varchar(255) DEFAULT NULL COMMENT '类型',
  `params` json DEFAULT NULL COMMENT '参数',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统事件';
CREATE TABLE `edgeSysLockers` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `key` varchar(255) DEFAULT NULL COMMENT '键值',
  `version` bigint(20) unsigned DEFAULT '0' COMMENT '版本号',
  `timeoutAt` bigint(11) unsigned DEFAULT '0' COMMENT '超时时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='并发锁';
CREATE TABLE `edgeSysSettings` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `code` varchar(255) DEFAULT NULL COMMENT '代号',
  `value` json DEFAULT NULL COMMENT '配置值',
  PRIMARY KEY (`id`),
  UNIQUE KEY `code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统配置';
CREATE TABLE `edgeTCPFirewallPolicies` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `adminId` int(11) DEFAULT NULL COMMENT '管理员ID',
  `userId` int(11) unsigned DEFAULT '0' COMMENT '用户ID',
  `templateId` int(11) unsigned DEFAULT '0' COMMENT '模版ID',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='TCP防火墙';
CREATE TABLE `edgeUsers` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `username` varchar(64) DEFAULT NULL COMMENT '用户名',
  `password` varchar(32) DEFAULT NULL COMMENT '密码',
  `fullname` varchar(64) DEFAULT NULL COMMENT '真实姓名',
  `createdAt` bigint(11) unsigned DEFAULT '0' COMMENT '创建时间',
  `updatedAt` bigint(11) unsigned DEFAULT '0' COMMENT '修改时间',
  `state` tinyint(1) unsigned DEFAULT '1' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户';
CREATE TABLE `edgeVersions` (
  `id` bigint(16) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `version` varchar(64) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='数据库结构版本';
