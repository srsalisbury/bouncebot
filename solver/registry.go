package solver

import "sync"

// Registry holds registered solvers.
type Registry struct {
	mu      sync.RWMutex
	solvers map[string]Solver
}

// NewRegistry creates a new solver registry.
func NewRegistry() *Registry {
	return &Registry{
		solvers: make(map[string]Solver),
	}
}

// Register adds a solver to the registry.
// If a solver with the same name already exists, it will be replaced.
func (r *Registry) Register(solver Solver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.solvers[solver.Name()] = solver
}

// Get returns a solver by name, or nil if not found.
func (r *Registry) Get(name string) Solver {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.solvers[name]
}

// All returns all registered solvers.
func (r *Registry) All() []Solver {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Solver, 0, len(r.solvers))
	for _, s := range r.solvers {
		result = append(result, s)
	}
	return result
}

// Names returns the names of all registered solvers.
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]string, 0, len(r.solvers))
	for name := range r.solvers {
		result = append(result, name)
	}
	return result
}

// DefaultRegistry is the global solver registry.
// Solvers can register themselves via init() functions.
var DefaultRegistry = NewRegistry()

// Register is a convenience function to register a solver with the default registry.
func Register(solver Solver) {
	DefaultRegistry.Register(solver)
}
