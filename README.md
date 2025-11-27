# 博客系统

## 一、用户表结构设计

models包中建立create_table.sql

```
CREATE TABLE 'user' (
    'id' bigint(20) NOT NULL AUTO_INCREMENT,
    'user_id' bigint(20) NOT NULL,
    'username' varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
    'password' varchar(64) COLLATE utf8mb4_general_ci NOT NULL,
    'email' varchar(64) COLLATE utf8mb4_general_ci,
    'gender' tinyint(4) NOT NULL DEFAULT '0',
    'create_time' timestamp NULL DEFAULT CURRENT_TIMESTAMP,
    'update_time' timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE
                CURRENT_TIMESTAMP,
    PRIMARY KEY ('id'),
    UNIQUE KEY 'idx_username' ('username') USING BTREE,
    UNIQUE KEY 'idx_user_id' ('user_id') USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
```

1. COLLATE utf8mb4_general_ci 表示该字段使用 utf8mb4 编码，并采用 general_ci 排序规则（大小写不敏感排序）
2. 'gender' tinyint(4) NOT NULL DEFAULT '0'
gender: 性别，类型为 tinyint(4)，通常表示为 0（未知）、1（男）、2（女）等数值。不能为 NULL，默认值为 0（表示未知）。
3. create_time: 记录创建时间，类型为 timestamp，允许为空（NULL）。如果插入记录时没有指定该字段，默认使用当前时间戳（CURRENT_TIMESTAMP）
4. ON UPDATE CURRENT_TIMESTAMP:每次更新记录时，自动更新该字段为当前时间
5. UNIQUE KEY 'idx_username' ('username') USING BTREE
定义 username 字段为唯一索引，确保每个用户名在表中是唯一的。使用 B+ 树（BTREE）作为索引结构,'idx_username'是索引名称
6. ENGINE=InnoDB: 表的存储引擎是 InnoDB，它支持事务、行级锁、外键等高级功能
7. DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci: 表的默认字符集为 utf8mb4（支持多字节字符，如 emoji），默认排序规则为 utf8mb4_general_ci（大小写不敏感排序）

