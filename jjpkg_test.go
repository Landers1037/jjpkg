/*
landers Apps
Author: landers
Github: github.com/landers1037
*/

package main

import (
	"testing"
)

func TestGetOS(t *testing.T) {
	b := checkGo()
	if !b {
		t.Error("can't find go")
	}
	t.Logf("check go %v", b)
}

func TestCreateVer(t *testing.T) {
	createVersionTag("1.0.0")
	t.Skip()
}

func TestCreateJJ(t *testing.T) {
	createOwnjj("app", "app_test", "1.0.0", "just for test")
	t.Skip()
}