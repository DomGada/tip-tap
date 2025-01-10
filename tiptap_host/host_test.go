package tiptaphost

import "testing"

func TestCreateHost(t *testing.T) {
	host, err := CreateHost("25565")
	if err != nil {
		t.Fatalf("TestCreateHost failed: %s\n", err.Error())
	}
	DeleteHost(host)
}
