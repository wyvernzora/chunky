package section

// Section represents a Markdown heading section and its associated body content.
// The root section is synthetic (Level = 0) and represents the entire file.
type Section struct {
	parent   *Section
	title    string
	level    int
	content  string
	children []*Section
}

// NewRoot creates a synthetic top-level section with the given title.
// Level is always 0. Caller is responsible for deriving the title.
func NewRoot(title string) *Section {
	return &Section{
		parent:   nil,
		title:    title,
		level:    0,
		children: make([]*Section, 0),
	}
}

// IsRoot reports whether this is the top-level synthetic section.
func (s *Section) IsRoot() bool {
	return s.parent == nil
}

// Parent returns the parent section (nil for root).
func (s *Section) Parent() *Section {
	return s.parent
}

// Title of the section.
func (s *Section) Title() string {
	return s.title
}

// Level is the Markdown heading depth (root = 0).
func (s *Section) Level() int {
	return s.level
}

// Content returns the accumulated Markdown body text.
func (s *Section) Content() string {
	return s.content
}

// SetContent replaces the body.
func (s *Section) SetContent(text string) { s.content = text }

// ResetContent clears the body.
func (s *Section) ResetContent() { s.content = "" }

// AppendContent appends to the body.
func (s *Section) AppendContent(fragment string) { s.content += fragment }

// PrependContent prepends to the body.
func (s *Section) PrependContent(fragment string) { s.content = fragment + s.content }

// Children returns a copy of the section's children.
func (s *Section) Children() []*Section {
	out := make([]*Section, len(s.children))
	copy(out, s.children)
	return out
}

// CreateChild adds a new section at the specified heading level with optional content.
func (s *Section) CreateChild(title string, level int, content string) *Section {
	child := &Section{
		parent:   s,
		title:    title,
		level:    level,
		content:  content,
		children: make([]*Section, 0),
	}
	s.children = append(s.children, child)
	return child
}
