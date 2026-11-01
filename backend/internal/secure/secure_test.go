package secure

import "testing"

func TestEncryptDecryptRoundTrip(t *testing.T) {
	if err := Init("test-secret"); err != nil {
		t.Fatalf("Init gagal: %v", err)
	}

	plain := `{"API_KEY":"rahasia","OTHER":"nilai"}`
	enc, err := Encrypt(plain)
	if err != nil {
		t.Fatalf("Encrypt gagal: %v", err)
	}
	if enc == plain {
		t.Fatalf("ciphertext sama dengan plaintext")
	}

	dec, err := Decrypt(enc)
	if err != nil {
		t.Fatalf("Decrypt gagal: %v", err)
	}
	if dec != plain {
		t.Fatalf("hasil dekripsi = %q, ingin %q", dec, plain)
	}
}

// Nilai lama (tanpa prefix versi) harus tetap terbaca sebagai plaintext.
func TestDecryptLegacyPlaintext(t *testing.T) {
	if err := Init("test-secret"); err != nil {
		t.Fatalf("Init gagal: %v", err)
	}
	got, err := Decrypt(`{"OLD":"1"}`)
	if err != nil {
		t.Fatalf("Decrypt legacy gagal: %v", err)
	}
	if got != `{"OLD":"1"}` {
		t.Fatalf("Decrypt legacy = %q", got)
	}
}

func TestHashIDAndSessionID(t *testing.T) {
	if HashID("abc") != HashID("abc") {
		t.Fatalf("HashID tidak deterministik")
	}
	if HashID("abc") == HashID("abd") {
		t.Fatalf("HashID berbeda input menghasilkan hash sama")
	}

	id, err := NewSessionID()
	if err != nil {
		t.Fatalf("NewSessionID gagal: %v", err)
	}
	if len(id) < 32 || id[:5] != "sess_" {
		t.Fatalf("format session id tidak sesuai: %q", id)
	}
	if other, _ := NewSessionID(); other == id {
		t.Fatalf("session id tidak unik")
	}
}
