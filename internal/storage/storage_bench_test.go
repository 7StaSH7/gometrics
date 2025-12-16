package storage

import (
	"testing"

	"github.com/7StaSH7/gometrics/internal/config"
)

func BenchmarkAddCounter(b *testing.B) {
	cfg := &config.ServerConfig{}
	s := NewStorage(cfg)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		s.Add("test_counter", 1)
	}
}

func BenchmarkReplaceGauge(b *testing.B) {
	cfg := &config.ServerConfig{}
	s := NewStorage(cfg)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		s.Replace("test_gauge", 123.45)
	}
}

func BenchmarkReadCounter(b *testing.B) {
	cfg := &config.ServerConfig{}
	s := NewStorage(cfg)
	s.Add("test_counter", 100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = s.ReadCounter("test_counter")
	}
}

func BenchmarkReadGauge(b *testing.B) {
	cfg := &config.ServerConfig{}
	s := NewStorage(cfg)
	s.Replace("test_gauge", 123.45)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = s.ReadGauge("test_gauge")
	}
}

func BenchmarkReadAll(b *testing.B) {
	cfg := &config.ServerConfig{}
	s := NewStorage(cfg)

	for i := 0; i < 100; i++ {
		s.Add("counter_"+string(rune('0'+i%10)), int64(i))
		s.Replace("gauge_"+string(rune('0'+i%10)), float64(i)*1.5)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = s.ReadAll()
	}
}
