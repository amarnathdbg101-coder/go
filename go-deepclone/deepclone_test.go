package deepclone

import (
	"reflect"
	"testing"
)

type Address struct {
	City   string
	Zip    string
	Coords *Coordinates
}

type Coordinates struct {
	Lat float64
	Lng float64
}

type User struct {
	ID        int
	Name      string
	Tags      []string
	Meta      map[string]any
	HomeAddr  *Address
	WorkAddr  *Address
	Emergency *User
}

func TestClone_StructWithPointersAndCollections(t *testing.T) {
	orig := User{
		ID:   1,
		Name: "Rahul",
		Tags: []string{"admin", "developer"},
		Meta: map[string]any{"tier": "gold", "logins": 42},
		HomeAddr: &Address{
			City: "Mumbai",
			Zip:  "400001",
			Coords: &Coordinates{
				Lat: 18.9220,
				Lng: 72.8347,
			},
		},
	}

	cloned := Clone(orig)

	// Verify deep equality initially
	if !reflect.DeepEqual(orig, cloned) {
		t.Fatalf("cloned object does not match original initially")
	}

	// Mutate clone's slice
	cloned.Tags[0] = "MUTATED"
	if orig.Tags[0] == "MUTATED" {
		t.Errorf("mutating clone's slice affected original")
	}

	// Mutate clone's map
	cloned.Meta["tier"] = "platinum"
	if orig.Meta["tier"] == "platinum" {
		t.Errorf("mutating clone's map affected original")
	}

	// Mutate clone's nested pointer
	cloned.HomeAddr.City = "Pune"
	if orig.HomeAddr.City == "Pune" {
		t.Errorf("mutating clone's nested pointer affected original")
	}

	// Verify pointers are distinct memory addresses
	if orig.HomeAddr == cloned.HomeAddr {
		t.Errorf("HomeAddr pointers are identical (shallow copy)")
	}
	if orig.HomeAddr.Coords == cloned.HomeAddr.Coords {
		t.Errorf("Coords pointers are identical (shallow copy)")
	}
}

func TestClone_CircularReference(t *testing.T) {
	type Node struct {
		Val  int
		Next *Node
	}

	n1 := &Node{Val: 1}
	n2 := &Node{Val: 2}
	n1.Next = n2
	n2.Next = n1 // Cycle: n1 -> n2 -> n1

	cloned := Clone(n1)

	if cloned.Val != 1 || cloned.Next.Val != 2 {
		t.Fatalf("cloned circular nodes have incorrect values")
	}

	// Verify cycle is preserved in the cloned graph
	if cloned.Next.Next != cloned {
		t.Errorf("cycle was broken in clone: cloned.Next.Next != cloned")
	}

	// Verify pointers are distinct from originals
	if cloned == n1 || cloned.Next == n2 {
		t.Errorf("cloned nodes point to original memory addresses")
	}
}

func TestClone_NilPointersAndSlices(t *testing.T) {
	var ptr *User = nil
	clonedPtr := Clone(ptr)
	if clonedPtr != nil {
		t.Errorf("expected nil pointer clone to be nil")
	}

	var slice []int = nil
	clonedSlice := Clone(slice)
	if clonedSlice != nil {
		t.Errorf("expected nil slice clone to be nil")
	}

	var m map[string]int = nil
	clonedMap := Clone(m)
	if clonedMap != nil {
		t.Errorf("expected nil map clone to be nil")
	}
}
