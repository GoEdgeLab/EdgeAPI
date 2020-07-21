package configs

import (
	_ "github.com/iwind/TeaGo/bootstrap"
	"testing"
)

func TestSharedAPIConfig(t *testing.T) {
	config, err := SharedAPIConfig()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(config)
}
