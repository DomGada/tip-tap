package host

import "testing"

func TestHostInputValid(t *testing.T) {

}

func TestCreateHost(t *testing.T) {
	host, err := StartHostConnection("25565")
	if err != nil {
		t.Fatalf("TestCreateHost failed: %s\n", err.Error())
	}
	StopHostConnection(host)
}
