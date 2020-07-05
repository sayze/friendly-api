// Package DB implements an in DB storage interface for all the entities.
package memory

import (
	"github.com/sayze/foundu-taker-api/internal/friend"
)

type DB struct {
	friends []*friend.Friend
}

func (db *DB) All() ([]*friend.Friend, error) {
	var activeFriends []*friend.Friend
	for _, fr := range db.friends {
		if fr.Active {
			activeFriends = append(activeFriends, fr)
		}
	}

	return activeFriends, nil
}

func (db *DB) GetFriend(id int64) (*friend.Friend, error) {
	return db.getFriendById(id), nil
}

func (db *DB) AddFriend(image string, name string) (*friend.Friend, error) {
	dbSize := int64(len(db.friends))
	db.friends = append(db.friends, &friend.Friend{
		ID:     dbSize + 1,
		Name:   name,
		Image:  image,
		Active: true,
	})
	return db.friends[dbSize], nil
}

func (db *DB) UpdateFriend(id int64, image string, name string) (*friend.Friend, error) {
	fr := db.getFriendById(id)

	if fr == nil {
		return nil, nil
	}

	fr.Name = name
	fr.Image = image
	return fr, nil
}

func (db *DB) DeleteFriend(id int64) (int, error) {
	fr := db.getFriendById(id)

	if fr == nil {
		return 0, nil
	}

	fr.Active = false

	return 1, nil
}

func (db *DB) getFriendById(id int64) *friend.Friend {
	for _, fr := range db.friends {
		if id == fr.ID && fr.Active {
			return fr
		}
	}

	return nil
}
