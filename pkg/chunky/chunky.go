package chunky

// Hello just returns a greeting from the library layer.
// This is here so pkg/ is importable right away.
func Hello(name string) string {
	if name == "" {
		name = "world"
	}
	return "chunky says hello, " + name + "!"
}
