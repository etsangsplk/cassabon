package logging

import (
	"testing"
)

func TestStats(t *testing.T) {

	err := S.Open("127.0.0.1:8125", "cassabon")
	if err != nil {
		t.Errorf("statsd: Unexpected error opening statsd client: %v", err)
	}
	S.Close()

	err = S.Open("999.0.0.1:8125", "cassabon")
	if err == nil {
		t.Errorf("statsd: No error reported when opening an invalid IP address")
	}
	S.Close()
}