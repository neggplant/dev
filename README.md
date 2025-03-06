优化建议
连接池：MongoDB 客户端默认使用连接池，可以通过 options.Client().SetMaxPoolSize() 调整连接池大小。

索引：为常用查询字段（如 email）创建索引，以提高查询性能。

事务：如果需要事务支持，MongoDB 4.0+ 支持多文档事务。

错误处理：统一处理 MongoDB 的错误，返回更友好的错误信息。
