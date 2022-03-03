package mapset

import (
	"testing"
)

func BenchmarkIterCast(b *testing.B) {
	s := NewThreadUnsafeSet()
	for i := 0; i < 5; i++ {
		s.Add(i)
	}

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		for range *(s.(*ThreadUnsafeSet)) {
		}
	}
}

func BenchmarkIter(b *testing.B) {
	s := NewThreadUnsafeSet()
	for i := 0; i < 5; i++ {
		s.Add(i)
	}

	b.ResetTimer()
	for j := 0; j < b.N; j++ {
		for range s.Iter() {
		}
	}
}
