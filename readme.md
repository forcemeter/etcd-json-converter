# etcd-json-converter

A simple ETCD export/import tool for backup and restore operations.

## Features

- Export ETCD data to JSON file
- Import JSON file to ETCD
- Support key prefix filtering
- Support limiting number of exported keys

## Usage

### Export

Export all keys from ETCD cluster:

```shell
etcd-json-converter export \
  --endpoint=10.10.0.3:2379,10.10.0.4:2379,10.10.0.5:2379 \
  --file=/tmp/output.json
```

Export with prefix and limit:

```shell
etcd-json-converter export \
  --endpoint=10.10.0.3:2379,10.10.0.4:2379,10.10.0.5:2379 \
  --prefix=/your/prefix \
  --limit=100 \
  --file=/tmp/output.json
```

### Import

Import data from JSON file to ETCD cluster:

```shell
etcd-json-converter import \
  --endpoint=10.10.0.3:2379,10.10.0.4:2379,10.10.0.5:2379 \
  --file=/tmp/input.json
```

### Status

Check ETCD cluster status:

```shell
etcd-json-converter status \
  --endpoint=10.10.0.3:2379
```

## Docker

### Quick Start

Start ETCD and build the tool:

```shell
docker compose up -d etcd
docker compose build
```

### Export with Docker

```shell
docker compose run --rm etcd-json-converter export \
  -e etcd:2379 \
  -f /data/backup.json
```

### Import with Docker

```shell
docker compose run --rm etcd-json-converter import \
  -e etcd:2379 \
  -f /data/backup.json
```

### Check Status

```shell
docker compose run --rm etcd-json-converter status -e etcd:2379
```

### Connect to External ETCD

```shell
docker compose run --rm etcd-json-converter export \
  -e 10.10.0.3:2379,10.10.0.4:2379 \
  -f /data/backup.json
```

> Note: Exported files are saved in `./data` directory.

## Tips

### Data Migration

You can modify the JSON file before importing for data migration:

```shell
# Replace domain names
sed -i 's/baidu.com/google.com/g' /tmp/input.json

# Replace IP addresses
sed -i 's/192.168.1.1/10.0.0.1/g' /tmp/input.json
```

## Options

| Option | Short | Description |
|--------|-------|-------------|
| `--endpoint` | `-e` | ETCD endpoints (comma-separated) |
| `--file` | `-f` | JSON file path (default: load.json) |
| `--prefix` | `-p` | Key prefix for export (default: /) |
| `--limit` | `-l` | Limit number of keys to export (default: all) |
