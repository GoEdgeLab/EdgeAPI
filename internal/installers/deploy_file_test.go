// Copyright 2021 Liuxiangchao iwind.liu@gmail.com. All rights reserved.

package installers

import (
	"io"
	"testing"
)

func TestDeployFile_Sum(t *testing.T) {
	d := &DeployFile{Path: "deploy_test.txt"}
	sum, err := d.Sum()
	if err != nil {
		t.Log("err:", err)
		return
	}
	t.Log("sum:", sum)
}

func TestDeployFile_Read(t *testing.T) {
	d := &DeployFile{Path: "deploy_test.txt"}

	var offset int64
	for i := 0; i < 3; i++ {
		data, newOffset, err := d.Read(offset)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Log("err: ", err)
			return
		}
		t.Log("offset:", newOffset, "data:", string(data))
		offset = newOffset
	}
}
