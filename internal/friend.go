package internal

// Friend implements the main entity.
type Friend struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Active bool   `json:"active"`
}

// FriendService defines a Friend implementation.
type FriendService interface {
	All() ([]*Friend, error)
	GetFriend(id int64) (*Friend, error)
	AddFriend(image string, name string) (*Friend, error)
	UpdateFriend(id int64, image string, name string) (*Friend, error)
	DeleteFriend(id int64) (int, error)
}
