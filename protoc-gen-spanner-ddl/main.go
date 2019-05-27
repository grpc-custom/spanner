package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	gogenerator "github.com/golang/protobuf/protoc-gen-go/generator"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	options "github.com/grpc-custom/spanner/proto"
	"github.com/grpc-custom/spanner/protoc-gen-spanner-ddl/ddl"
	"github.com/grpc-custom/spanner/protoc-gen-spanner-ddl/descriptor"
)

func parseParameter(parameter string) {
	if parameter == "" {
		return
	}
	for _, p := range strings.Split(parameter, ",") {
		spec := strings.SplitN(p, "=", 2)
		if len(spec) == 1 {
			if err := flag.CommandLine.Set(spec[0], ""); err != nil {
				glog.Fatalf("Cannot set flag %s", p)
			}
			continue
		}
		name, value := spec[0], spec[1]
		if strings.HasPrefix(name, "M") {
			continue
		}
		if err := flag.CommandLine.Set(name, value); err != nil {
			glog.Fatalf("Cannot set flag %s", p)
		}
	}
}

func emitFiles(out []*plugin.CodeGeneratorResponse_File) {
	emitResp(&plugin.CodeGeneratorResponse{
		File: out,
	})
}

func emitError(err error) {
	emitResp(&plugin.CodeGeneratorResponse{Error: proto.String(err.Error())})
}

func emitResp(resp *plugin.CodeGeneratorResponse) {
	buf, err := proto.Marshal(resp)
	if err != nil {
		glog.Fatal(err)
	}
	if _, err := os.Stdout.Write(buf); err != nil {
		glog.Fatal(err)
	}
}

func main() {
	flag.Parse()
	defer glog.Flush()

	req, err := descriptor.ParseRequest(os.Stdin)
	if err != nil {
		glog.Fatal(err)
	}

	parseParameter(req.GetParameter())

	reg := descriptor.NewRegistry()
	if err := reg.Load(req); err != nil {
		emitError(err)
		return
	}

	database := ddl.New()
	for _, path := range req.FileToGenerate {
		target, err := reg.LookupFile(path)
		if err != nil {
			glog.Fatal(err)
		}

		for _, msg := range target.Messages {
			if !proto.HasExtension(msg.Options, options.E_Schema) {
				continue
			}
			ext, _ := proto.GetExtension(msg.Options, options.E_Schema)
			opts, ok := ext.(*options.Schema)
			if !ok {
				continue
			}

			indexes := make([]*ddl.Index, 0, len(opts.Index))
			for _, idx := range opts.Index {
				index := &ddl.Index{
					Table:        opts.Table,
					Name:         idx.Name,
					Columns:      idx.Columns,
					Interleave:   idx.Interleave,
					NullFiltered: idx.NullFiltered,
					Storing:      idx.Storing,
					Unique:       idx.Unique,
				}
				indexes = append(indexes, index)
			}

			columns := make([]*ddl.Column, 0, len(msg.Fields))
			for _, field := range msg.Fields {
				column := &ddl.Column{}
				columns = append(columns, column)

				name := gogenerator.CamelCase(field.GetName())
				column.Name = name
				column.SetType(field)

				if !proto.HasExtension(field.Options, options.E_Column) {
					continue
				}
				ext, _ := proto.GetExtension(field.Options, options.E_Column)
				opts, ok := ext.(*options.Column)
				if !ok {
					continue
				}

				if opts.Name != "" {
					column.Name = opts.Name
				}
				column.Length = int(opts.Length)
				column.Nullable = opts.Nullable
				column.AllowCommitTimestamp = opts.AllowCommitTimestamp
			}

			for _, imp := range opts.Import {
				f, err := reg.LookupFile(imp.Path)
				if err != nil {
					glog.Fatal(err)
				}
				for _, msg := range f.Messages {
					if msg.GetName() != imp.Type {
						continue
					}
					for _, field := range msg.Fields {
						column := &ddl.Column{}
						columns = append(columns, column)

						name := gogenerator.CamelCase(field.GetName())
						column.Name = name
						column.SetType(field)

						if !proto.HasExtension(field.Options, options.E_Column) {
							continue
						}
						ext, _ := proto.GetExtension(field.Options, options.E_Column)
						opts, ok := ext.(*options.Column)
						if !ok {
							continue
						}

						if opts.Name != "" {
							column.Name = opts.Name
						}
						column.Length = int(opts.Length)
						column.Nullable = opts.Nullable
						column.AllowCommitTimestamp = opts.AllowCommitTimestamp
					}
				}
			}

			table := &ddl.Table{
				Name:       opts.Table,
				PrimaryKey: opts.PrimaryKey,
				Indexes:    indexes,
				Columns:    columns,
			}
			if opts.Interleave != nil {
				table.Interleave = &ddl.Interleave{
					Name: opts.Interleave.Name,
				}
				if opts.Interleave.OnDelete != options.OnDelete_NONE {
					table.Interleave.OnDelete = int(opts.Interleave.OnDelete)
				}
			}
			database.Tables = append(database.Tables, table)
		}
	}

	files := make([]*plugin.CodeGeneratorResponse_File, 0)
	for _, table := range database.Tables {
		output := fmt.Sprintf("table_%s.ddl", table.Name)
		f := &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(output),
			Content: proto.String(table.String()),
		}
		files = append(files, f)
		glog.Info(table.String())

		buf := &strings.Builder{}
		for _, idx := range table.Indexes {
			buf.WriteString(idx.String())
			buf.WriteString("\n")
			glog.Info(idx.String())
		}
		output = fmt.Sprintf("index_%s.ddl", table.Name)
		f = &plugin.CodeGeneratorResponse_File{
			Name:    proto.String(output),
			Content: proto.String(buf.String()),
		}
		files = append(files, f)
	}

	emitFiles(files)
}
