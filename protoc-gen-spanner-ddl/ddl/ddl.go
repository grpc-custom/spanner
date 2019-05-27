package ddl

import (
	"fmt"
	"strings"

	godescriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/grpc-custom/spanner/protoc-gen-spanner-ddl/descriptor"
)

const (
	timestampProto = ".google.protobuf.Timestamp"
	dateProto      = ".google.type.Date"
)

type ColumnType struct {
	Type       SchemaType
	IsRepeated bool
}

func (c ColumnType) String() string {
	typ := c.getType()
	if c.IsRepeated {
		typ = fmt.Sprintf("ARRAY<%s>", typ)
	}
	return typ
}

func (c ColumnType) getType() string {
	switch c.Type {
	case Int64Type:
		return "INT64"
	case Float64Type:
		return "FLOAT64"
	case StringType:
		return "STRING"
	case BoolType:
		return "BOOL"
	case BytesType:
		return "BYTES"
	case TimestampType:
		return "TIMESTAMP"
	case DateType:
		return "DATE"
	}
	return ""
}

type SchemaType int

const (
	UnknownType SchemaType = iota
	ArrayType
	BoolType
	BytesType
	DateType
	Float64Type
	Int64Type
	StringType
	StructType
	TimestampType
)

type Table struct {
	Name       string
	PrimaryKey string
	Interleave *Interleave
	Columns    []*Column
	Indexes    []*Index
}

func (t *Table) String() string {
	buf := &strings.Builder{}
	buf.Grow(200)
	buf.WriteString("CREATE TABLE `")
	buf.WriteString(t.Name)
	buf.WriteString("` (\n")

	for _, column := range t.Columns {
		buf.WriteString("\t`" + column.Name + "`")
		buf.WriteString(" " + column.Type.String())
		if column.Type.Type == StringType || column.Type.Type == BytesType {
			if column.Length <= 0 {
				buf.WriteString("(MAX)")
			} else {
				buf.WriteString(fmt.Sprintf("(%d)", column.Length))
			}
		}
		if !column.Nullable {
			buf.WriteString(" NOT NULL")
		}
		if column.AllowCommitTimestamp {
			buf.WriteString(fmt.Sprintf(" OPTIONS (allow_commit_timestamp=%t)", column.AllowCommitTimestamp))
		}
		buf.WriteString(",\n")
	}

	buf.WriteString(")")

	if t.PrimaryKey != "" {
		buf.WriteString(" PRIMARY KEY (")
		buf.WriteString(t.PrimaryKey)
		buf.WriteString(")")
	}
	if t.Interleave != nil {
		buf.WriteString(",\n")
		buf.WriteString("\tINTERLEAVE IN PARENT ")
		buf.WriteString("`" + t.Interleave.Name + "`")
		if t.Interleave.OnDelete == 1 {
			buf.WriteString(" ON DELETE CASCADE")
		}
		if t.Interleave.OnDelete == 2 {
			buf.WriteString(" ON DELETE ON ACTION")
		}
	}

	return buf.String()
}

type Interleave struct {
	Name     string
	OnDelete int
}

type Column struct {
	Name                 string
	Type                 ColumnType
	Length               int
	Nullable             bool
	AllowCommitTimestamp bool
}

func (c *Column) SetType(field *descriptor.Field) {
	c.Type.Type = c.getType(field)
	c.Type.IsRepeated = field.GetLabel() == godescriptor.FieldDescriptorProto_LABEL_REPEATED
}

func (c *Column) getType(field *descriptor.Field) SchemaType {
	switch field.GetType() {
	case
		godescriptor.FieldDescriptorProto_TYPE_INT32,
		godescriptor.FieldDescriptorProto_TYPE_INT64,
		godescriptor.FieldDescriptorProto_TYPE_UINT32,
		godescriptor.FieldDescriptorProto_TYPE_UINT64,
		godescriptor.FieldDescriptorProto_TYPE_SINT32,
		godescriptor.FieldDescriptorProto_TYPE_SINT64,
		godescriptor.FieldDescriptorProto_TYPE_FIXED32,
		godescriptor.FieldDescriptorProto_TYPE_FIXED64,
		godescriptor.FieldDescriptorProto_TYPE_SFIXED32,
		godescriptor.FieldDescriptorProto_TYPE_SFIXED64:
		return Int64Type
	case
		godescriptor.FieldDescriptorProto_TYPE_DOUBLE,
		godescriptor.FieldDescriptorProto_TYPE_FLOAT:
		return Float64Type
	case
		godescriptor.FieldDescriptorProto_TYPE_STRING:
		return StringType
	case
		godescriptor.FieldDescriptorProto_TYPE_BOOL:
		return BoolType
	case
		godescriptor.FieldDescriptorProto_TYPE_BYTES:
		return BytesType
	case
		godescriptor.FieldDescriptorProto_TYPE_MESSAGE:
		if field.GetTypeName() == timestampProto {
			return TimestampType
		}
		if field.GetTypeName() == dateProto {
			return DateType
		}
	}
	return UnknownType
}

type Index struct {
	Table        string
	Name         string
	Columns      string
	Interleave   string
	Storing      string
	NullFiltered bool
	Unique       bool
}

func (i *Index) String() string {
	buf := &strings.Builder{}
	buf.Grow(200)
	buf.WriteString("CREATE")
	if i.Unique {
		buf.WriteString(" UNIQUE")
	}
	if i.NullFiltered {
		buf.WriteString(" NULL_FILTERED")
	}
	buf.WriteString(" INDEX ")
	buf.WriteString("`" + i.Name + "`")
	buf.WriteString(" ON ")
	buf.WriteString("`" + i.Table + "`")
	buf.WriteString(fmt.Sprintf(" (\n\t%s\n)", i.Columns))
	if i.Storing != "" {
		buf.WriteString(fmt.Sprintf(" STORING (\n\t%s\n)", i.Storing))
	}
	if i.Interleave != "" {
		buf.WriteString(fmt.Sprintf(" INTERLEAVE IN %s", i.Interleave))
	}
	return buf.String()
}

type DDL struct {
	Tables []*Table
}

func New() *DDL {
	return &DDL{
		Tables: make([]*Table, 0),
	}
}

func (d *DDL) CreateTables() {
}
