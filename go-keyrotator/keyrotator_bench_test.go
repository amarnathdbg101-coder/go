package keyrotator

import "testing"

func BenchmarkRotator_GetNextKey(b *testing.B) {
	rotator := New([]string{
		"AIzaSyKey1",
		"AIzaSyKey2",
		"AIzaSyKey3",
		"AIzaSyKey4",
		"AIzaSyKey5",
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = rotator.GetNextKey()
	}
}

func BenchmarkRotator_GetNextKey_Parallel(b *testing.B) {
	rotator := New([]string{
		"AIzaSyKey1",
		"AIzaSyKey2",
		"AIzaSyKey3",
		"AIzaSyKey4",
		"AIzaSyKey5",
	})

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = rotator.GetNextKey()
		}
	})
}
