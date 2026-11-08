package godi

import (
	"errors"
	"testing"
)

type Database struct {
	URL    string
	Closed bool
}

func (d *Database) Close() error {
	d.Closed = true
	return nil
}

type UserRepository struct {
	DB *Database
}

type UserService struct {
	Repo *UserRepository
}

func TestContainer_SingletonAndResolution(t *testing.T) {
	c := New()

	RegisterSingleton(c, func(cont *Container) (*Database, error) {
		return &Database{URL: "postgres://localhost:5432/test"}, nil
	})

	RegisterSingleton(c, func(cont *Container) (*UserRepository, error) {
		db, err := Resolve[*Database](cont)
		if err != nil {
			return nil, err
		}
		return &UserRepository{DB: db}, nil
	})

	RegisterSingleton(c, func(cont *Container) (*UserService, error) {
		repo, err := Resolve[*UserRepository](cont)
		if err != nil {
			return nil, err
		}
		return &UserService{Repo: repo}, nil
	})

	svc1, err := Resolve[*UserService](c)
	if err != nil {
		t.Fatalf("failed to resolve UserService: %v", err)
	}

	svc2, err := Resolve[*UserService](c)
	if err != nil {
		t.Fatalf("failed to resolve UserService second time: %v", err)
	}

	// Singleton pointer equality
	if svc1 != svc2 {
		t.Errorf("expected identical singleton pointers")
	}
	if svc1.Repo.DB != svc2.Repo.DB {
		t.Errorf("expected identical DB singleton pointers")
	}

	// Cleanup
	db := svc1.Repo.DB
	if err := c.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if !db.Closed {
		t.Errorf("expected DB to be closed by container")
	}
}

func TestContainer_Transient(t *testing.T) {
	c := New()

	type RequestContext struct {
		ID int
	}

	counter := 0
	RegisterTransient(c, func(cont *Container) (*RequestContext, error) {
		counter++
		return &RequestContext{ID: counter}, nil
	})

	r1 := MustResolve[*RequestContext](c)
	r2 := MustResolve[*RequestContext](c)

	if r1 == r2 {
		t.Errorf("transient should return distinct instances")
	}
	if r1.ID == r2.ID {
		t.Errorf("expected distinct counter IDs")
	}
}

func TestContainer_NamedServices(t *testing.T) {
	c := New()

	RegisterNamedSingleton(c, "primary_db", func(cont *Container) (string, error) {
		return "primary_connection", nil
	})

	RegisterNamedSingleton(c, "replica_db", func(cont *Container) (string, error) {
		return "replica_connection", nil
	})

	prim, err := ResolveNamed[string](c, "primary_db")
	if err != nil || prim != "primary_connection" {
		t.Errorf("expected primary_connection, got %v", prim)
	}

	rep, err := ResolveNamed[string](c, "replica_db")
	if err != nil || rep != "replica_connection" {
		t.Errorf("expected replica_connection, got %v", rep)
	}
}

func TestContainer_CircularDependency(t *testing.T) {
	c := New()

	type ServiceA struct{ B any }
	type ServiceB struct{ A any }

	RegisterSingleton(c, func(cont *Container) (*ServiceA, error) {
		b, err := Resolve[*ServiceB](cont)
		if err != nil {
			return nil, err
		}
		return &ServiceA{B: b}, nil
	})

	RegisterSingleton(c, func(cont *Container) (*ServiceB, error) {
		a, err := Resolve[*ServiceA](cont)
		if err != nil {
			return nil, err
		}
		return &ServiceB{A: a}, nil
	})

	_, err := Resolve[*ServiceA](c)
	if err == nil || !errors.Is(err, ErrCircularDependency) {
		t.Errorf("expected ErrCircularDependency, got: %v", err)
	}
}

func TestContainer_ServiceNotFound(t *testing.T) {
	c := New()

	type NonExistent struct{}
	_, err := Resolve[*NonExistent](c)
	if err == nil || !errors.Is(err, ErrServiceNotFound) {
		t.Errorf("expected ErrServiceNotFound, got: %v", err)
	}
}
