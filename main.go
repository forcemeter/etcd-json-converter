package main

import (
	"context"
	"errors"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	"go.etcd.io/etcd/client/v3"
	"os"
	"strings"
	"time"
)

var (
	methodImport  = "import"
	methodExport  = "export"
	methodDefault = ""

	argEndpoint = []gcmd.Argument{
		gcmd.Argument{
			Name:   `endpoint`,
			Short:  `i`,
			Brief:  `ip:port,ip:port,ip:port`,
			Orphan: false,
		},
		gcmd.Argument{
			Name:   `file`,
			Short:  `f`,
			Brief:  `/tmp/load.json`,
			Orphan: true,
		},
		gcmd.Argument{
			Name:   `prefix`,
			Short:  `p`,
			Brief:  `export key prefix, must start with /, default /`,
			Orphan: true,
		},
		gcmd.Argument{
			Name:   `limit`,
			Short:  `l`,
			Brief:  `export file limit, default all`,
			Orphan: true,
		},
	}
	Main = &gcmd.Command{
		Name:        "etcd json converter",
		Brief:       "ETCD 简单导出导入工具",
		Description: "ETCD 简单导出导入工具",
	}
	ExportMethod = &gcmd.Command{
		Name:        methodExport,
		Brief:       "ETCD 导出数据",
		Description: "export data from etcd cluster and save as json file",
		Arguments:   argEndpoint,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			methodDefault = methodExport
			run(ctx, parser)
			return
		},
	}
	ImportMethod = &gcmd.Command{
		Name:        methodImport,
		Brief:       "ETCD 导入数据",
		Description: "import json to etcd cluster",
		Arguments:   argEndpoint,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			methodDefault = methodImport
			run(ctx, parser)
			return
		},
	}
	StatusMethod = &gcmd.Command{
		Name:        "status",
		Brief:       "ETCD 集群状态",
		Description: "show etcd cluster information",
		Arguments:   argEndpoint,
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			methodDefault = "status"
			run(ctx, parser)
			return
		},
	}
)

var (
	config clientv3.Config
	client *clientv3.Client
	err    error
)

func main() {
	err := Main.AddCommand(ExportMethod, ImportMethod, StatusMethod)
	if err != nil {
		panic(err)
	}
	Main.Run(gctx.New())
}

func run(ctx context.Context, parser *gcmd.Parser) {
	endpoint := parser.GetOpt("endpoint").String()
	if g.IsEmpty(endpoint) {
		glog.Info(ctx, "--endpoint 参数不能为空")
		os.Exit(0)
	}

	file := parser.GetOpt("file").String()
	if g.IsEmpty(file) {
		file = "load.json"
	}

	limit := parser.GetOpt("limit").Int64()
	prefix := parser.GetOpt("prefix").String()

	endpoints := gstr.Explode(",", endpoint)
	glog.Info(ctx, endpoints)
	config = clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		glog.Panic(ctx, err)
	}
	defer client.Close()

	switch methodDefault {
	case methodExport:
		if err := Export(ctx, prefix, limit, file); err != nil {
			glog.Error(ctx, err)
		}
	case methodImport:
		if err := Import(ctx, file); err != nil {
			glog.Error(ctx, err)
		}
	default:
		Status(ctx)
	}
}

func Import(ctx context.Context, file string) error {
	if !gfile.IsFile(file) {
		return errors.New("文件不存在:" + file)
	}

	data := map[string]string{}
	err := gjson.DecodeTo(gfile.GetContents(file), &data)
	if err != nil {
		glog.Fatal(ctx, err)
	}
	for k, v := range data {
		glog.Info(ctx, k)
		r, err := client.Put(ctx, k, v)
		if err != nil {
			glog.Error(ctx, "写入失败")
			return err
		} else {
			glog.Info(ctx, k, r.Header.String())
		}
	}

	return nil
}

func Export(ctx context.Context, prefix string, limit int64, file string) error {
	if g.IsEmpty(prefix) || !gstr.HasPrefix(prefix, "/") {
		prefix = "/"
	}

	if limit < 0 {
		limit = 0
	}

	glog.Infof(ctx, "export %d rows with prefix %s", limit, prefix)

	response, err := client.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithLimit(limit))
	if err != nil {
		return err
	}

	data := map[string]string{}
	for _, kv := range response.Kvs {
		data[string(kv.Key)] = string(kv.Value)
	}

	gfile.PutContents(file, gjson.MustEncodeString(data))
	glog.Infof(ctx, "已保存到 %s", gfile.RealPath(file))
	glog.Infof(ctx, "共导出数据 %d 条", len(data))

	return nil
}

func Status(ctx context.Context) {
	sts, err := client.Status(ctx, strings.Join(config.Endpoints, ","))
	glog.Info(ctx, sts, err)
}
