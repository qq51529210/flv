package flv

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestFLV(t *testing.T) {
	dir := filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "qq51529210", "flv")
	// 源文件
	src, err := os.Open(filepath.Join(dir, "1.flv"))
	if nil != err {
		t.Fatal(err)
	}
	defer src.Close()
	// 目标文件
	dst, err := os.OpenFile(filepath.Join(dir, "2.flv"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if nil != err {
		t.Fatal(err)
	}
	defer dst.Close()
	//
	var header Header
	_, err = header.ReadFrom(src)
	if nil != err {
		t.Fatal(err)
	}
	_, err = header.WriteTo(dst)
	if nil != err {
		t.Fatal(err)
	}
	var tag Tag
	for {
		_, err = tag.ReadFrom(src)
		if nil != err {
			if err == io.EOF {
				break
			}
			t.Fatal(err)
		}
		_, err = tag.WriteTo(dst)
		if nil != err {
			t.Fatal(err)
		}
	}
}
