package inmem_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/facebookgo/inmem"
)

var src = rand.NewSource(time.Now().UnixNano())

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	keyLen        = 10
	valLen        = 100
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func benchmarkManyThings(b *testing.B, c inmem.Cache) {

	var key = RandString(keyLen)

	// cache miss
	actual, found := c.Get(key)
	if found == true || actual != nil {
		b.Fail()
	}

	// add the value
	val := RandString(valLen)
	c.Add(key, val, time.Now().Add(time.Hour))

	// cache hit
	actual, found = c.Get(key)
	if found == false || actual == nil || val != actual {
		b.Fail()
	}

	// replace value
	val = RandString(valLen)
	c.Add(key, val, time.Now().Add(time.Hour))

	// check for replace value
	actual, found = c.Get(key)
	if found == false || actual == nil || val != actual {
		b.Fail()
	}

	// replace with expired
	val = RandString(valLen)
	c.Add(key, val, time.Now().Add(-1*time.Hour))

	// try to get the expired key
	actual, found = c.Get(key)
	if found == true || actual != nil {
		b.Fail()
	}

	// readd the value so we can remove it
	c.Add(key, val, time.Now().Add(time.Hour))

	// remove the key
	c.Remove(key)

	// try to get the empty key
	actual, found = c.Get(key)
	if found == true || actual != nil {
		b.Fail()
	}

	// readd the original
	c.Add(key, val, time.Now().Add(time.Hour))
}

func benchmarkLockedCache(cacheSize int, b *testing.B) {
	cache := inmem.NewLocked(cacheSize)
	for n := 0; n < b.N; n++ {
		benchmarkManyThings(b, cache)
	}
}

func benchmarkUnlockedCache(cacheSize int, b *testing.B) {
	cache := inmem.NewUnlocked(cacheSize)
	for n := 0; n < b.N; n++ {
		benchmarkManyThings(b, cache)
	}
}

func BenchmarkSmallUnlocked(b *testing.B) {
	benchmarkUnlockedCache(10, b)
}

func BenchmarkSmallLocked(b *testing.B) {
	benchmarkLockedCache(10, b)
}

func BenchmarkLargeUnlocked(b *testing.B) {
	benchmarkUnlockedCache(1000, b)
}

func BenchmarkLargeLocked(b *testing.B) {
	benchmarkUnlockedCache(1000, b)
}

func BenchmarkHugeUnlocked(b *testing.B) {
	benchmarkUnlockedCache(100000, b)
}

func BenchmarkHugeLocked(b *testing.B) {
	benchmarkLockedCache(100000, b)
}
