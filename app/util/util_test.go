package util_test

import (
	"regexp"
	"testing"

	"github.com/ryan-willis/qotd/app/util"
)

func TestGenerateRoomCode(t *testing.T) {
	exp := regexp.MustCompile("^[" + util.LETTERS + "]{4}$")
	for i := 0; i < 100; i++ {
		code := util.GenerateRoomCode()
		if len(code) != 4 {
			t.Errorf("Expected code to be 4 characters long, got %d", len(code))
		}
		if code == "" {
			t.Errorf("Expected code to not be empty")
		}
		if !exp.Match([]byte(code)) {
			t.Errorf("Expected code to match regex, got %s", code)
		}
		t.Logf("Generated code: %s\n", code)
	}
}
