package entity

// FriendStore describes storage interface.
type FriendStore interface {
	CreateFriend(Friend) (*Friend, error)
	ViewFriend(int64) (*Friend, error)
	DeleteFriend(int64) (bool, error)
	UpdateFriend(*FriendUpdate) (*Friend, error)
}

// Friend describes the environment entity.
type Friend struct {
	ID     int64
	Name   string
	Age    int
	Active bool
}

// FriendUpdate defines fields required for an update.
type FriendUpdate struct {
	ID   int64
	Name string
	Age  int
}
