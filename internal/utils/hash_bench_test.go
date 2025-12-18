package utils

import (
	"testing"
)

func BenchmarkGenerateSHA256(b *testing.B) {
	value := `{"id":"test_counter","type":"counter","delta":100}`
	key := "secret-key"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GenerateSHA256(value, key)
	}
}

func BenchmarkVerifySHA256(b *testing.B) {
	value := `{"id":"test_counter","type":"counter","delta":100}`
	key := "secret-key"
	hash := GenerateSHA256(value, key)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = VerifySHA256(hash, hash)
	}
}

func BenchmarkGenerateSHA256LargePayload(b *testing.B) {
	value := `[{"id":"counter_1","type":"counter","delta":100},{"id":"counter_2","type":"counter","delta":200},{"id":"gauge_1","type":"gauge","value":123.45},{"id":"gauge_2","type":"gauge","value":678.90}]`
	key := "secret-key"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GenerateSHA256(value, key)
	}
}

func BenchmarkGenerateSHA256Bytes(b *testing.B) {
	value := []byte(`{"id":"test_counter","type":"counter","delta":100}`)
	key := "secret-key"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GenerateSHA256Bytes(value, key)
	}
}

func BenchmarkGenerateSHA256BytesLargePayload(b *testing.B) {
	value := []byte(`[{"id":"counter_1","type":"counter","delta":100},{"id":"counter_2","type":"counter","delta":200},{"id":"gauge_1","type":"gauge","value":123.45},{"id":"gauge_2","type":"gauge","value":678.90}]`)
	key := "secret-key"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = GenerateSHA256Bytes(value, key)
	}
}
