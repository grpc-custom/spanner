package descriptor

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/golang/glog"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
)

type Registry struct {
	msgs       map[string]*Message
	files      map[string]*File
	pkgs       map[string]string
	importPath string
	prefix     string
}

func NewRegistry() *Registry {
	return &Registry{
		msgs:  make(map[string]*Message),
		files: make(map[string]*File),
		pkgs:  make(map[string]string),
	}
}

func (r *Registry) Load(req *plugin.CodeGeneratorRequest) error {
	for _, file := range req.GetProtoFile() {
		r.loadFile(file)
	}
	return nil
}

func (r *Registry) loadFile(file *descriptor.FileDescriptorProto) {
	pkg := &GoPackage{
		Path: r.goPackagePath(file),
		Name: r.defaultGoPackageName(file),
	}
	f := &File{
		FileDescriptorProto: file,
		GoPkg:               pkg,
	}
	r.files[file.GetName()] = f
	r.registerMsg(f, nil, file.GetMessageType())
}

func (r *Registry) registerMsg(file *File, outerPath []string, msgs []*descriptor.DescriptorProto) {
	for i, msg := range msgs {
		m := &Message{
			File:            file,
			Outers:          outerPath,
			DescriptorProto: msg,
			Index:           i,
		}
		for _, fd := range msg.GetField() {
			m.Fields = append(m.Fields, &Field{
				Message:              m,
				FieldDescriptorProto: fd,
			})
		}
		file.Messages = append(file.Messages, m)
		r.msgs[m.FQMN()] = m
		glog.V(1).Infof("register name: %s", m.FQMN())

		var outers []string
		outers = append(outers, outerPath...)
		outers = append(outers, m.GetName())
		r.registerMsg(file, outers, m.GetNestedType())
	}
}

func (r *Registry) goPackagePath(f *descriptor.FileDescriptorProto) string {
	name := f.GetName()
	if pkg, ok := r.pkgs[name]; ok {
		return path.Join(r.prefix, pkg)
	}

	gopkg := f.Options.GetGoPackage()
	idx := strings.LastIndex(gopkg, "/")
	if idx >= 0 {
		if sc := strings.LastIndex(gopkg, ";"); sc > 0 {
			gopkg = gopkg[:sc+1-1]
		}
		return gopkg
	}

	return path.Join(r.prefix, path.Dir(name))
}

func (r *Registry) defaultGoPackageName(f *descriptor.FileDescriptorProto) string {
	name := r.packageIdentityName(f)
	return sanitizePackageName(name)
}

func (r *Registry) packageIdentityName(f *descriptor.FileDescriptorProto) string {
	if f.Options != nil && f.Options.GoPackage != nil {
		gopkg := f.Options.GetGoPackage()
		idx := strings.LastIndex(gopkg, "/")
		if idx < 0 {
			gopkg = gopkg[idx+1:]
		}

		gopkg = gopkg[idx+1:]
		// package name is overrided with the string after the
		// ';' character
		sc := strings.IndexByte(gopkg, ';')
		if sc < 0 {
			return sanitizePackageName(gopkg)

		}
		return sanitizePackageName(gopkg[sc+1:])
	}
	if p := r.importPath; len(p) != 0 {
		if i := strings.LastIndex(p, "/"); i >= 0 {
			p = p[i+1:]
		}
		return p
	}

	if f.Package == nil {
		base := filepath.Base(f.GetName())
		ext := filepath.Ext(base)
		return strings.TrimSuffix(base, ext)
	}
	return f.GetPackage()
}

func sanitizePackageName(pkgName string) string {
	pkgName = strings.Replace(pkgName, ".", "_", -1)
	pkgName = strings.Replace(pkgName, "-", "_", -1)
	return pkgName
}

func (r *Registry) LookupFile(name string) (*File, error) {
	f, ok := r.files[name]
	if !ok {
		return nil, fmt.Errorf("no such file given: %s", name)
	}
	return f, nil
}

func (r *Registry) LookupMsg(location, name string) (*Message, error) {
	glog.V(1).Infof("lookup %s from %s", name, location)
	if strings.HasPrefix(name, ".") {
		m, ok := r.msgs[name]
		if !ok {
			return nil, fmt.Errorf("no message found: %s", name)
		}
		return m, nil
	}

	if !strings.HasPrefix(location, ".") {
		location = fmt.Sprintf(".%s", location)
	}
	components := strings.Split(location, ".")
	for len(components) > 0 {
		fqmn := strings.Join(append(components, name), ".")
		if m, ok := r.msgs[fqmn]; ok {
			return m, nil
		}
		components = components[:len(components)-1]
	}
	return nil, fmt.Errorf("no message found: %s", name)
}
