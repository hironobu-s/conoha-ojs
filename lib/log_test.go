package lib

import (
	"testing"
)

func TestGetLogInstance(t *testing.T) {
	log1 := GetLogInstance()
	if log1 == nil {
		t.Errorf("log shoud not be nil.")
	}

	log2 := GetLogInstance()
	if log1 != log2 {
		t.Errorf("log1 and log2 is not same struct.")
	}
}
