SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for global_config
-- ----------------------------
DROP TABLE IF EXISTS `global_config`;
CREATE TABLE `global_config` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID（固定为1）',
  `http_addr` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT ':80' COMMENT 'HTTP 监听地址',
  `https_addr` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT ':443' COMMENT 'HTTPS 监听地址',
  `max_body_size` bigint DEFAULT '5242880' COMMENT '最大请求体大小（字节），默认5MB',
  `max_header_mb` int DEFAULT '1' COMMENT '最大请求头大小（MB），默认1MB',
  `cert_pem` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SSL 证书 PEM 文本',
  `key_pem` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SSL 私钥 PEM 文本',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci ROW_FORMAT=DYNAMIC COMMENT='全局配置表（仅一条记录）';

-- ----------------------------
-- Records of global_config
-- ----------------------------
BEGIN;
INSERT INTO `global_config` (`id`, `http_addr`, `https_addr`, `max_body_size`, `max_header_mb`, `cert_pem`, `key_pem`) VALUES (1, ':80', ':443', 5242880, 1, 'xxxx', 'xxxx');
COMMIT;

-- ----------------------------
-- Table structure for server
-- ----------------------------
DROP TABLE IF EXISTS `server`;
CREATE TABLE `server` (
  `id` int NOT NULL AUTO_INCREMENT,
  `server_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '域名',
  `enable_https` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否启用https访问',
  `active_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'v1' COMMENT '激活版本',
  `gray_version` varchar(64) DEFAULT NULL COMMENT '灰度版本',
  `gray_header_key` varchar(255) DEFAULT NULL COMMENT '灰度Key',
  `gray_header_value` varchar(255) DEFAULT NULL COMMENT '灰度Value',
  PRIMARY KEY (`id`),
  UNIQUE KEY `server_name` (`server_name`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of server
-- ----------------------------
BEGIN;
INSERT INTO `server` (`id`, `server_name`, `enable_https`, `active_version`, `gray_version`, `gray_header_key`, `gray_header_value`) VALUES (1, 'cutejava.odboy.cn', 1, '20260412105706', 'v3', 'env', 'gray');
INSERT INTO `server` (`id`, `server_name`, `enable_https`, `active_version`, `gray_version`, `gray_header_key`, `gray_header_value`) VALUES (2, 'kenaito-dns.odboy.com', 0, 'v1', NULL, NULL, NULL);
COMMIT;

-- ----------------------------
-- Table structure for server_version
-- ----------------------------
DROP TABLE IF EXISTS `server_version`;
CREATE TABLE `server_version` (
  `id` int NOT NULL AUTO_INCREMENT,
  `server_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '域名',
  `version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '版本号',
  `bucket_path` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '在Bucket中的路径',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_server_version` (`server_name`,`version`),
  CONSTRAINT `server_version_ibfk_1` FOREIGN KEY (`server_name`) REFERENCES `server` (`server_name`) ON DELETE CASCADE
) ENGINE=InnoDB AUTO_INCREMENT=9 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- ----------------------------
-- Records of server_version
-- ----------------------------
BEGIN;
INSERT INTO `server_version` (`id`, `server_name`, `version`, `bucket_path`) VALUES (1, 'cutejava.odboy.cn', 'v1', '/cutejava/v1');
INSERT INTO `server_version` (`id`, `server_name`, `version`, `bucket_path`) VALUES (2, 'cutejava.odboy.cn', 'v2', '/cutejava/v2');
INSERT INTO `server_version` (`id`, `server_name`, `version`, `bucket_path`) VALUES (3, 'cutejava.odboy.cn', 'v3', '/cutejava/v3');
INSERT INTO `server_version` (`id`, `server_name`, `version`, `bucket_path`) VALUES (5, 'kenaito-dns.odboy.com', 'v1', '/kenaito-dns/v1');
INSERT INTO `server_version` (`id`, `server_name`, `version`, `bucket_path`) VALUES (6, 'cutejava.odboy.cn', 'v4', '/cutejava/v4');
INSERT INTO `server_version` (`id`, `server_name`, `version`, `bucket_path`) VALUES (7, 'cutejava.odboy.cn', '20260412104442', 'cutejava-front/20260412104442');
INSERT INTO `server_version` (`id`, `server_name`, `version`, `bucket_path`) VALUES (8, 'cutejava.odboy.cn', '20260412105706', 'cutejava-front/20260412105706');
COMMIT;

SET FOREIGN_KEY_CHECKS = 1;
