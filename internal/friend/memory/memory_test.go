package memory

import (
	"github.com/sayze/foundu-taker-api/internal/friend"
	"github.com/stretchr/testify/assert"
	"testing"
)

func seedDB() *DB {
	data := []*friend.Friend{
		&friend.Friend{
			ID:     1,
			Name:   "Adam Smith",
			Image:  "fake1",
			Active: true,
		},
		&friend.Friend{
			ID:     121,
			Name:   "Nolan Andrew",
			Image:  "fake23",
			Active: true,
		},
		&friend.Friend{
			ID:     31,
			Name:   "Russel Evans",
			Image:  "",
			Active: false,
		},
	}
	return &DB{friends: data}
}

func TestDB_AddFriend(t *testing.T) {
	var db DB
	newFriend, err := db.AddFriend("fake-image", "Adam Smith")
	assert.Nil(t, err)
	assert.Equal(t, &friend.Friend{
		ID:     1,
		Name:   "Adam Smith",
		Image:  "fake-image",
		Active: true,
	}, newFriend)
}

func TestDB_GetFriend(t *testing.T) {
	db := seedDB()
	fr := db.getFriendById(121)
	assert.Equal(t, int64(121), fr.ID)
	assert.Equal(t, "Nolan Andrew", fr.Name)

	// Assert an inactive friend can't be accessed.
	assert.Nil(t, db.getFriendById(31))
}

func TestDB_DeleteFriend(t *testing.T) {
	db := seedDB()
	ct, err := db.DeleteFriend(121)
	assert.Nil(t, err)
	assert.Equal(t, 1, ct)
}

func TestDB_UpdateFriend(t *testing.T) {
	db := seedDB()
	fr, err := db.UpdateFriend(1, "new-image", "New Name")
	assert.Nil(t, err)
	assert.Equal(t, "new-image", fr.Image)
	assert.Equal(t, "New Name", fr.Name)
}
