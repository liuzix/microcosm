package resource

import "testing"

func TestManager(t *testing.T) {
	m := &Manager{
		resourceMap: map[string]string{},
	}
	m.AddResources("eID", []string{"rID1", "rID2"})
}
