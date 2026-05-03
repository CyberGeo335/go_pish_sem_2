package order

// Order describes an order stored inside order-service.
type Order struct {
	ID     int64   `json:"id"`
	UserID int64   `json:"user_id"`
	Item   string  `json:"item"`
	Price  float64 `json:"price"`
}

// UserDTO is a copy of the public user contract returned by user-service.
// order-service should not import the internal user-service code directly.
type UserDTO struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// OrderWithUser is an aggregated response built by order-service.
type OrderWithUser struct {
	Order Order   `json:"order"`
	User  UserDTO `json:"user"`
}
