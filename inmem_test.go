package inmem_test

import (
	"testing"
	"time"

	"github.com/facebookgo/ensure"
	"github.com/facebookgo/inmem"
)

func testManyThings(t *testing.T, c inmem.Cache) {
	const (
		k  = 1
		v1 = 2
		v2 = 3
	)

	// it's empty
	ensure.DeepEqual(t, c.Len(), 0)

	// not there to start with
	actual, found := c.Get(k)
	ensure.False(t, found)
	ensure.Nil(t, actual)

	// add it
	c.Add(k, v1, time.Now().Add(time.Hour))

	// now it's there
	actual, found = c.Get(k)
	ensure.True(t, found)
	ensure.DeepEqual(t, actual, v1)

	// we have some items
	ensure.DeepEqual(t, c.Len(), 1)

	// replace it
	c.Add(k, v2, time.Now().Add(time.Hour))

	// now find the new value
	actual, found = c.Get(k)
	ensure.True(t, found)
	ensure.DeepEqual(t, actual, v2)

	// we still only have 1 item
	ensure.DeepEqual(t, c.Len(), 1)

	// remove it
	c.Remove(k)

	// not there any more
	actual, found = c.Get(k)
	ensure.False(t, found)
	ensure.Nil(t, actual)

	// it's empty again
	ensure.DeepEqual(t, c.Len(), 0)
}

func TestManyThingsUnlocked(t *testing.T) {
	testManyThings(t, inmem.NewUnlocked(10))
}

func TestManyThingsLocked(t *testing.T) {
	testManyThings(t, inmem.NewLocked(10))
}

func TestPanicNewUnlockedSizeZero(t *testing.T) {
	defer ensure.PanicDeepEqual(t, "inmem: must provide a positive size")
	_ = inmem.NewUnlocked(0)
}

func TestPanicNewLockedSizeZero(t *testing.T) {
	defer ensure.PanicDeepEqual(t, "inmem: must provide a positive size")
	_ = inmem.NewLocked(0)
}

func TestCacheSize(t *testing.T) {
	c := inmem.NewUnlocked(2)
	e := time.Now().Add(time.Hour)
	c.Add(1, 1, e)
	c.Add(2, 2, e)
	c.Add(3, 3, e)
	ensure.DeepEqual(t, c.Len(), 2)
	_, found := c.Get(1)
	ensure.False(t, found)
}

func TestTTLExpired(t *testing.T) {
	c := inmem.NewUnlocked(2)
	e := time.Now().Add(-time.Hour)
	c.Add(1, 1, e)
	ensure.DeepEqual(t, c.Len(), 1)
	_, found := c.Get(1)
	ensure.False(t, found)
	ensure.DeepEqual(t, c.Len(), 0)
}
