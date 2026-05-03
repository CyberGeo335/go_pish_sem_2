package student

// Student describes a simple student entity returned by the demo API.
type Student struct {
	ID       int64  `json:"id"`
	FullName string `json:"full_name"`
	Group    string `json:"group"`
	Email    string `json:"email"`
}
