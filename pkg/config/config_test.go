package config_test

import (
	"strings"
	"testing"

	"github.com/stuart-warren/localpod/pkg/config"
)

func TestGoodEnum(t *testing.T) {
	cfg := strings.NewReader(`{
			"image": "someimage",
			"shutdownAction": "none"
		}`)
	dc, err := config.DevContainerFromFile(cfg)
	if err != nil {
		t.Errorf("error decoding json: %s", err)
	}
	if dc.ShutdownAction != config.None {
		t.Errorf("expected config.None: %v", dc.ShutdownAction)
	}
}

func TestBadEnum(t *testing.T) {
	cfg := strings.NewReader(`{
			"image": "someimage",
			"shutdownAction": "unexpected"
		}`)
	_, err := config.DevContainerFromFile(cfg)
	if err == nil {
		t.Errorf("expected error")
	}
}
