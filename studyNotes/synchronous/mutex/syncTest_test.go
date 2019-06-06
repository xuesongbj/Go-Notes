package syncTest

import (
    "testing"
)

func BenchmarkUseChan(b *testing.B) {
    for n := 0; n < b.N; n++ {
        UseChan()
    }
}

func BenchmarkUseMutex(b *testing.B) {
    for n := 0; n < b.N; n++ {
        UseMutex()
    }
}
