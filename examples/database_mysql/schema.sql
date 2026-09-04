-- ==============================================================
-- Godeniter 2.0 MySQL 示例数据表结构与初始数据
-- ==============================================================

CREATE DATABASE IF NOT EXISTS `godeniter_demo` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `godeniter_demo`;

-- 1. 文章主表
CREATE TABLE IF NOT EXISTS `articles` (
    `id` INT AUTO_INCREMENT PRIMARY KEY COMMENT '文章主键ID',
    `title` VARCHAR(255) NOT NULL COMMENT '文章标题',
    `content` TEXT NOT NULL COMMENT '文章内容正文',
    `author` VARCHAR(100) NOT NULL DEFAULT '匿名' COMMENT '作者',
    `views` INT NOT NULL DEFAULT 0 COMMENT '浏览量计数',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1正常 0软删除/禁用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    INDEX `idx_status_views` (`status`, `views`),
    INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章资讯表';

-- 2. 灌入初始演示数据
INSERT INTO `articles` (`title`, `content`, `author`, `views`, `status`) VALUES
('Godeniter 2.0 极速上手', '基于纯 Go 标准库与 MySQL 打造的高性能轻量 Web 框架，支持 CodeIgniter 风格 ActiveRecord。', 'admin', 25, 1),
('MySQL 生产连接池与索引调优', '生产环境中建议合理配置 MaxOpenConns 与 MaxIdleConns，避免频繁短连接握手。', 'xbt', 99, 1),
('Go 依赖注入与洋葱圈中间件', '请求级上下文 Injector 使得控制器方法可以声明式获取数据库连接与服务实例。', 'dev_team', 56, 1);
