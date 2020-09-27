package neon

// moduler : Creates interface for input modules
type Moduler interface {
	// Placeholder so we donot expose empty interface
	placeholder()
}

// Module : Module is default implementation for moduler interface
// modules can embed this structure so can create new module without implementing moduler
type Module struct {
}

// Default implementaion for placeholder
func (m Module) placeholder() {}
