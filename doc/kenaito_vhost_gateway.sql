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

 Date: 11/04/2026 09:00:33
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
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '全局配置表（仅一条记录）' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of global_config
-- ----------------------------
INSERT INTO `global_config` VALUES (1, ':80', ':443', 5242880, 1, 'xx', 'xx');

-- ----------------------------
-- Table structure for servers
-- ----------------------------
DROP TABLE IF EXISTS `servers`;
CREATE TABLE `servers`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `server_name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '域名（server_name）',
  `root` varchar(1024) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL COMMENT '静态文件根目录路径',
  `enable_https` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否启用 HTTPS（0: 否, 1: 是）',
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_server_name`(`server_name` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '虚拟主机配置表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Records of servers
-- ----------------------------
INSERT INTO `servers` VALUES (1, 'cutejava.odboy.com', 'E:\\DevFiles\\dist\\cutejava-front', 1);
INSERT INTO `servers` VALUES (2, 'kenaito-dns.odboy.cn', 'E:\\DevFiles\\dist\\kenaito-dns-front', 0);

SET FOREIGN_KEY_CHECKS = 1;
