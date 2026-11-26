package id

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-7[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

// New menghasilkan UUID versi 7 (time-ordered): 48 bit timestamp milidetik,
// diikuti bit acak. Tidak butuh dependency eksternal.
func New() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand gagal hanya pada kondisi sistem yang sangat rusak;
		// tetap hasilkan nilai unik dari waktu agar tidak panic.
		now := time.Now().UnixNano()
		for i := 0; i < 8; i++ {
			b[i] = byte(now >> (8 * i))
		}
	}

	ms := uint64(time.Now().UnixMilli())
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)

	b[6] = (b[6] & 0x0f) | 0x70 // versi 7
	b[8] = (b[8] & 0x3f) | 0x80 // varian RFC 4122

	hexed := hex.EncodeToString(b[:])
	return fmt.Sprintf("%s-%s-%s-%s-%s", hexed[0:8], hexed[8:12], hexed[12:16], hexed[16:20], hexed[20:32])
}

// Valid memeriksa apakah string adalah UUID v7 yang valid.
func Valid(s string) bool {
	return uuidPattern.MatchString(strings.ToLower(strings.TrimSpace(s)))
}
