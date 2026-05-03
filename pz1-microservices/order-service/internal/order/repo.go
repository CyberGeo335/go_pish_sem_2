package order

import (
	"errors"
	"sort"
)

var ErrOrderNotFound = errors.New("order not found")

// Repo is a simple in-memory order repository for the practice task.
type Repo struct {
	data map[int64]Order
}

func NewRepo() *Repo {
	return &Repo{
		data: map[int64]Order{
			101: {ID: 101, UserID: 1, Item: "Ноутбук", Price: 79990},
			102: {ID: 102, UserID: 2, Item: "Мышь", Price: 2490},
			103: {ID: 103, UserID: 1, Item: "Клавиатура", Price: 5990},
		},
	}
}

func (r *Repo) GetByID(id int64) (Order, error) {
	o, ok := r.data[id]
	if !ok {
		return Order{}, ErrOrderNotFound
	}

	return o, nil
}

func (r *Repo) GetByUserID(userID int64) []Order {
	orders := make([]Order, 0)
	for _, order := range r.data {
		if order.UserID == userID {
			orders = append(orders, order)
		}
	}

	sort.Slice(orders, func(i, j int) bool { return orders[i].ID < orders[j].ID })
	return orders
}
