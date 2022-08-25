## etcd-json-converter 一个简单的 ETCD 导出、导入工具，最简单的备份和还原工具

#### 备份，只需要单节点执行
```shell
etcd-tool export \
--endpoint=10.10.0.3:2379,10.10.0.4:2379,10.10.0.5:2379 \
--file=/tmp/output.json
```

```shell
etcd-tool export \
--endpoint=10.10.0.3:2379,10.10.0.4:2379,10.10.0.5:2379 \
--limit=0 --prefix=/your/prefix \
--file=/tmp/output.json
```

#### 还原，只需要单节点执行
```shell
etcd-tool import \
--endpoint=10.10.0.3:2379,10.10.0.4:2379,10.10.0.5:2379 \
--file=/tmp/input.json
```

#### 可以在导入前对文件进行替换，方便进行迁移
```shell
sed -i 's/baidu.com/google.com/g' /tmp/input.json
```