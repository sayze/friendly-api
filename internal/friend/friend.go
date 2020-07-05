package friend

// Friend implements the main entity.
type Friend struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Image  string `json:"image"`
	Active bool   `json:"active"`
}

type Service interface {
	All() ([]*Friend, error)
	GetFriend(id int64) (*Friend, error)
	AddFriend(image string, name string) (*Friend, error)
	UpdateFriend(id int64, image string, name string) (*Friend, error)
	DeleteFriend(id int64) (int, error)
}

// FriendUpdate defines fields required for an update.
type FriendUpdate struct {
	ID   int64
	Name string
	Age  int
}
