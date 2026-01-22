package main

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	"github.com/gogf/gf/v2/text/gstr"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultDialTimeout = 5 * time.Second
	defaultFile        = "load.json"
	defaultPrefix      = "/"
)

var commonArgs = []gcmd.Argument{
	{
		Name:   "endpoint",
		Short:  "e",
		Brief:  "etcd endpoints (e.g., ip:port,ip:port)",
		Orphan: false,
	},
	{
		Name:   "file",
		Short:  "f",
		Brief:  "json file path (default: load.json)",
		Orphan: true,
	},
	{
		Name:   "prefix",
		Short:  "p",
		Brief:  "key prefix for export, must start with / (default: /)",
		Orphan: true,
	},
	{
		Name:   "limit",
		Short:  "l",
		Brief:  "limit number of keys to export (default: all)",
		Orphan: true,
	},
}

func main() {
	mainCmd := &gcmd.Command{
		Name:        "etcd-json-converter",
		Brief:       "ETCD Simple Export/Import Tool",
		Description: "A simple tool to export/import data between ETCD and JSON files",
	}

	exportCmd := &gcmd.Command{
		Name:        "export",
		Brief:       "Export data from ETCD to JSON file",
		Description: "Export data from etcd cluster and save as JSON file",
		Arguments:   commonArgs,
		Func:        runExport,
	}

	importCmd := &gcmd.Command{
		Name:        "import",
		Brief:       "Import data from JSON file to ETCD",
		Description: "Import JSON data to etcd cluster",
		Arguments:   commonArgs,
		Func:        runImport,
	}

	statusCmd := &gcmd.Command{
		Name:        "status",
		Brief:       "Show ETCD cluster status",
		Description: "Display etcd cluster information",
		Arguments:   commonArgs,
		Func:        runStatus,
	}

	if err := mainCmd.AddCommand(exportCmd, importCmd, statusCmd); err != nil {
		panic(err)
	}
	mainCmd.Run(gctx.New())
}

func newClient(ctx context.Context, endpoint string) (*clientv3.Client, error) {
	if g.IsEmpty(endpoint) {
		return nil, errors.New("--endpoint is required")
	}

	endpoints := gstr.Explode(",", endpoint)
	glog.Infof(ctx, "connecting to endpoints: %v", endpoints)

	return clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: defaultDialTimeout,
	})
}

func runExport(ctx context.Context, parser *gcmd.Parser) error {
	client, err := newClient(ctx, parser.GetOpt("endpoint").String())
	if err != nil {
		glog.Error(ctx, err)
		return err
	}
	defer client.Close()

	file := parser.GetOpt("file").String()
	if g.IsEmpty(file) {
		glog.Info(ctx, "use default file:", defaultFile)
		file = defaultFile
	}

	// Replace "time" placeholder with current datetime
	if strings.Contains(file, "-time") {
		file = strings.ReplaceAll(file, "-time", time.Now().Format("-20060102-150405"))
		glog.Info(ctx, "use time format file:", file)
	}

	prefix := parser.GetOpt("prefix").String()
	if g.IsEmpty(prefix) || !gstr.HasPrefix(prefix, "/") {
		prefix = defaultPrefix
	}

	limit := parser.GetOpt("limit").Int64()
	if limit < 0 {
		limit = 0
	}

	return exportData(ctx, client, prefix, limit, file)
}

func runImport(ctx context.Context, parser *gcmd.Parser) error {
	client, err := newClient(ctx, parser.GetOpt("endpoint").String())
	if err != nil {
		glog.Error(ctx, err)
		return err
	}
	defer client.Close()

	file := parser.GetOpt("file").String()
	if g.IsEmpty(file) {
		glog.Info(ctx, "use default file:", defaultFile)
		file = defaultFile
	}

	return importData(ctx, client, file)
}

func runStatus(ctx context.Context, parser *gcmd.Parser) error {
	endpoint := parser.GetOpt("endpoint").String()
	client, err := newClient(ctx, endpoint)
	if err != nil {
		glog.Error(ctx, err)
		return err
	}
	defer client.Close()

	status, err := client.Status(ctx, strings.Split(endpoint, ",")[0])
	if err != nil {
		glog.Errorf(ctx, "failed to get status: %v", err)
		return err
	}

	glog.Infof(ctx, "cluster status: %+v", status)
	return nil
}

func importData(ctx context.Context, client *clientv3.Client, file string) error {
	if !gfile.IsFile(file) {
		return errors.New("file not found: " + file)
	}

	content := gfile.GetContents(file)
	data := make(map[string]string)
	if err := gjson.DecodeTo(content, &data); err != nil {
		return err
	}

	glog.Infof(ctx, "importing %d keys from %s", len(data), file)

	for key, value := range data {
		resp, err := client.Put(ctx, key, value)
		if err != nil {
			glog.Errorf(ctx, "failed to write key %s: %v", key, err)
			return err
		}
		glog.Infof(ctx, "imported: %s (revision: %d)", key, resp.Header.Revision)
	}

	glog.Infof(ctx, "successfully imported %d keys", len(data))
	return nil
}

func exportData(ctx context.Context, client *clientv3.Client, prefix string, limit int64, file string) error {
	glog.Infof(ctx, "exporting keys with prefix '%s' (limit: %d)", prefix, limit)

	opts := []clientv3.OpOption{clientv3.WithPrefix()}
	if limit > 0 {
		opts = append(opts, clientv3.WithLimit(limit))
	}

	response, err := client.Get(ctx, prefix, opts...)
	if err != nil {
		return err
	}

	data := make(map[string]string, len(response.Kvs))
	for _, kv := range response.Kvs {
		data[string(kv.Key)] = string(kv.Value)
	}

	if err := gfile.PutContents(file, gjson.MustEncodeString(data)); err != nil {
		return err
	}

	glog.Infof(ctx, "saved to %s", gfile.RealPath(file))
	glog.Infof(ctx, "exported %d keys", len(data))
	return nil
}
