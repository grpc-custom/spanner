package descriptor

import (
	"fmt"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type GoPackage struct {
	Path  string
	Name  string
	Alias string
}

func (p *GoPackage) Standard() bool {
	return !strings.Contains(p.Path, ".")
}

func (p *GoPackage) String() string {
	if p.Alias == "" {
		return fmt.Sprintf("%q", p.Path)
	}
	return fmt.Sprintf("%s %q", p.Alias, p.Path)
}

type File struct {
	*descriptor.FileDescriptorProto
	GoPkg    *GoPackage
	Messages []*Message
}

type Message struct {
	*descriptor.DescriptorProto
	File   *File
	Outers []string
	Fields []*Field
	Index  int
}

func (m *Message) FQMN() string {
	components := []string{""}
	if m.File.Package != nil {
		components = append(components, m.File.GetPackage())
	}
	components = append(components, m.Outers...)
	components = append(components, m.GetName())
	return strings.Join(components, ".")
}

type Field struct {
	*descriptor.FieldDescriptorProto
	Message      *Message
	FieldMessage *Message
}
