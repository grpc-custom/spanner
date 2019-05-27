package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	if os.Getenv("RUN_AS_PROTOC_GEN_GO") != "" {
		main()
		os.Exit(0)
	}
}

func TestGolden(t *testing.T) {
	workdir, err := ioutil.TempDir("", "protoc-gen-spanner-ddl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(workdir)

	packages := make(map[string][]string)
	if err := filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if !strings.HasSuffix(path, ".proto") {
			return nil
		}
		dir := filepath.Dir(path)
		packages[dir] = append(packages[dir], path)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, sources := range packages {
		args := []string{
			"-I=testdata",
			"-I=" + os.Getenv("GOPATH") + "/src",
			"-I=" + os.Getenv("GOPATH") + "/src/github.com/googleapis/googleapis:.",
			"--spanner-ddl_out=logtostderr=true,v=1:" + workdir,
		}
		args = append(args, sources...)
		protoc(t, args)
	}
}

func protoc(t *testing.T, args []string) {
	cmd := exec.Command("protoc", "--plugin=protoc-gen-spanner-ddl="+os.Args[0])
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = append(os.Environ(), "RUN_AS_PROTOC_GEN_GO=1")
	out, err := cmd.CombinedOutput()
	if len(out) > 0 || err != nil {
		t.Log("RUNNING: ", strings.Join(cmd.Args, " "))
	}
	if len(out) > 0 {
		t.Log(string(out))
	}
	if err != nil {
		t.Fatalf("protoc: %v", err)
	}
}
