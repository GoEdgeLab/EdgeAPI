package models

import (
	"testing"
)

func TestDBNodeInitializer_loop(t *testing.T) {
	initializer := NewDBNodeInitializer()
	err := initializer.loop()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(accessLogDBMapping), len(httpAccessLogDAOMapping))
}
