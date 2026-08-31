package main

import (
	"path/filepath"
	"testing"
)

// tempHome은 테스트용 임시 홈 디렉터리를 설정합니다
func tempHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	return home
}
