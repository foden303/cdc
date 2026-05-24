package registry

import (
	"sync"
	"testing"
)

func TestTableRegistry_RegisterAndLookup(t *testing.T) {
	r := NewTableRegistry()

	// Initially not registered
	_, active := r.Lookup("src-1", "public", "users")
	if active {
		t.Fatal("expected table to not be active before registration")
	}

	// Register
	r.Register("src-1", "public", "users", 4)
	partitionCount, active := r.Lookup("src-1", "public", "users")
	if !active {
		t.Fatal("expected table to be active after registration")
	}
	if partitionCount != 4 {
		t.Fatalf("expected partitionCount=4, got %d", partitionCount)
	}
}

func TestTableRegistry_MaxPartitionCount(t *testing.T) {
	r := NewTableRegistry()

	r.Register("src-1", "public", "users", 4)
	r.Register("src-1", "public", "users", 8)

	partitionCount, active := r.Lookup("src-1", "public", "users")
	if !active {
		t.Fatal("expected table to be active")
	}
	if partitionCount != 8 {
		t.Fatalf("expected partitionCount=8 (max), got %d", partitionCount)
	}

	// Lower value should not reduce partition count
	r.Register("src-1", "public", "users", 2)
	partitionCount, _ = r.Lookup("src-1", "public", "users")
	if partitionCount != 8 {
		t.Fatalf("expected partitionCount=8 (unchanged), got %d", partitionCount)
	}
}

func TestTableRegistry_RefCount(t *testing.T) {
	r := NewTableRegistry()

	// Register twice
	r.Register("src-1", "public", "users", 4)
	r.Register("src-1", "public", "users", 4)

	// Unregister once — should still be active
	r.Unregister("src-1", "public", "users")
	_, active := r.Lookup("src-1", "public", "users")
	if !active {
		t.Fatal("expected table to still be active with refCount=1")
	}

	// Unregister again — should be removed
	r.Unregister("src-1", "public", "users")
	_, active = r.Lookup("src-1", "public", "users")
	if active {
		t.Fatal("expected table to be inactive after all refs removed")
	}
}

func TestTableRegistry_DifferentSources(t *testing.T) {
	r := NewTableRegistry()

	r.Register("src-1", "public", "users", 4)
	r.Register("src-2", "public", "users", 8)

	pc1, active1 := r.Lookup("src-1", "public", "users")
	pc2, active2 := r.Lookup("src-2", "public", "users")

	if !active1 || !active2 {
		t.Fatal("expected both sources to be active")
	}
	if pc1 != 4 {
		t.Fatalf("expected src-1 partitionCount=4, got %d", pc1)
	}
	if pc2 != 8 {
		t.Fatalf("expected src-2 partitionCount=8, got %d", pc2)
	}

	// Unregister src-1 should not affect src-2
	r.Unregister("src-1", "public", "users")
	_, active1 = r.Lookup("src-1", "public", "users")
	_, active2 = r.Lookup("src-2", "public", "users")
	if active1 {
		t.Fatal("expected src-1 to be inactive")
	}
	if !active2 {
		t.Fatal("expected src-2 to still be active")
	}
}

func TestTableRegistry_UnregisterNonExistent(t *testing.T) {
	r := NewTableRegistry()

	// Should not panic
	r.Unregister("src-1", "public", "users")

	_, active := r.Lookup("src-1", "public", "users")
	if active {
		t.Fatal("expected table to not be active")
	}
}

func TestTableRegistry_ConcurrentAccess(t *testing.T) {
	r := NewTableRegistry()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Register("src-1", "public", "users", 4)
		}()
	}
	wg.Wait()

	_, active := r.Lookup("src-1", "public", "users")
	if !active {
		t.Fatal("expected table to be active after concurrent registrations")
	}

	// Concurrent unregistrations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Unregister("src-1", "public", "users")
		}()
	}
	wg.Wait()

	_, active = r.Lookup("src-1", "public", "users")
	if active {
		t.Fatal("expected table to be inactive after all concurrent unregistrations")
	}
}
