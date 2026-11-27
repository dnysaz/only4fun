package server

import "testing"

func TestEnvHashIsDeterministicAndSensitive(t *testing.T) {
	first := map[string]string{"DATABASE_URL": "a", "API_KEY": "b"}
	reordered := map[string]string{"API_KEY": "b", "DATABASE_URL": "a"}
	changed := map[string]string{"DATABASE_URL": "c", "API_KEY": "b"}

	if envHash(first) != envHash(reordered) {
		t.Fatal("env hash should not depend on map iteration order")
	}
	if envHash(first) == envHash(changed) {
		t.Fatal("env hash should change when a value changes")
	}
	if envHash(nil) == envHash(first) {
		t.Fatal("empty env should have a distinct hash")
	}
}
