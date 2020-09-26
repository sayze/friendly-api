package memory

import (
	"testing"

	"github.com/sayze/friendly-api/entity"
	"github.com/stretchr/testify/assert"
)

func seedDB() *FriendService {
	data := []*entity.Friend{
		{
			ID:    1,
			Name:  "Adam Smith",
			Image: "fake1",
		},
		{
			ID:    121,
			Name:  "Nolan Andrew",
			Image: "fake23",
		},
		{
			ID:    31,
			Name:  "Russel Evans",
			Image: "",
		},
	}
	return &FriendService{DB: data}
}

func TestDB_AddFriend(t *testing.T) {
	var db FriendService
	newFriend, err := db.AddFriend("fake-image", "Adam Smith")
	assert.Nil(t, err)
	assert.Equal(t, &entity.Friend{
		ID:    1,
		Name:  "Adam Smith",
		Image: "fake-image",
	}, newFriend)
}

func TestDB_GetFriend(t *testing.T) {
	db := seedDB()
	fr := db.getFriendById(121)
	assert.Equal(t, int64(121), fr.ID)
	assert.Equal(t, "Nolan Andrew", fr.Name)
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

func TestDB_All(t *testing.T) {
	db := seedDB()
	data, _ := db.All("")
	assert.Len(t, data, 3)
	data, _ = db.All("adam")
	assert.Len(t, data, 1)
}
