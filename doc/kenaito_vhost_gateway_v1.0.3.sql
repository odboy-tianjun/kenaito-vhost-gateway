/*
 Navicat Premium Dump SQL

 Source Server         : tianyiyun-mysql
 Source Server Type    : MySQL
 Source Server Version : 80025 (8.0.25)
 Source Host           : 219.151.187.115:3306
 Source Schema         : kenaito_vhost_gateway

 Target Server Type    : MySQL
 Target Server Version : 80025 (8.0.25)
 File Encoding         : 65001

 Date: 11/04/2026 10:50:59
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for global_config
-- ----------------------------
DROP TABLE IF EXISTS `global_config`;
CREATE TABLE `global_config`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID（固定为1）',
  `http_addr` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT ':80' COMMENT 'HTTP 监听地址',
  `https_addr` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT ':443' COMMENT 'HTTPS 监听地址',
  `max_body_size` bigint NULL DEFAULT 5242880 COMMENT '最大请求体大小（字节），默认5MB',
  `max_header_mb` int NULL DEFAULT 1 COMMENT '最大请求头大小（MB），默认1MB',
  `cert_pem` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SSL 证书 PEM 文本',
  `key_pem` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT 'SSL 私钥 PEM 文本',
  `minio_endpoint` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `minio_access_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `minio_secret_key` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  `minio_use_ssl` tinyint(1) NOT NULL DEFAULT 0,
  `minio_bucket` varchar(128) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '全局配置表（仅一条记录）' ROW_FORMAT = DYNAMIC;

-- ----------------------------
-- Records of global_config
-- ----------------------------
INSERT INTO `global_config` VALUES (1, ':80', ':443', 5242880, 1, 'asdasd', 'asdasd', '192.168.235.128:9000', 'root', 'kd123456789', 0, 'web-static');

-- ----------------------------
-- Table structure for server
-- ----------------------------
DROP TABLE IF EXISTS `server`;
CREATE TABLE `server`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `server_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `active_version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT 'v1',
  `enable_https` tinyint(1) NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `server_name`(`server_name` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of server
-- ----------------------------
INSERT INTO `server` VALUES (1, 'cutejava.odboy.cn', 'v2', 1);
INSERT INTO `server` VALUES (2, 'kenaito-dns.odboy.com', 'v1', 0);

-- ----------------------------
-- Table structure for server_version
-- ----------------------------
DROP TABLE IF EXISTS `server_version`;
CREATE TABLE `server_version`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `server_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `version` varchar(64) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `bucket_path` varchar(512) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_server_version`(`server_name` ASC, `version` ASC) USING BTREE,
  CONSTRAINT `server_version_ibfk_1` FOREIGN KEY (`server_name`) REFERENCES `server` (`server_name`) ON DELETE CASCADE ON UPDATE RESTRICT
) ENGINE = InnoDB AUTO_INCREMENT = 6 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of server_version
-- ----------------------------
INSERT INTO `server_version` VALUES (1, 'cutejava.odboy.cn', 'v1', '/cutejava/v1');
INSERT INTO `server_version` VALUES (2, 'cutejava.odboy.cn', 'v2', '/cutejava/v2');
INSERT INTO `server_version` VALUES (3, 'cutejava.odboy.cn', 'v3', '/cutejava/v3');
INSERT INTO `server_version` VALUES (5, 'kenaito-dns.odboy.com', 'v1', '/kenaito-dns/v1');

SET FOREIGN_KEY_CHECKS = 1;
