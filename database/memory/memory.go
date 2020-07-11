// Package DB implements an in DB storage interface for all the entities.
package memory

import (
	"github.com/sayze/friendly-api/entity"
)

type FriendService struct {
	DB []*entity.Friend
}

func NewService() *FriendService {
	return &FriendService{}
}

func (s *FriendService) All() ([]*entity.Friend, error) {
	var activeFriends []*entity.Friend
	for _, fr := range s.DB {
		if fr.Active {
			activeFriends = append(activeFriends, fr)
		}
	}

	return activeFriends, nil
}

func (s *FriendService) GetFriend(id int64) (*entity.Friend, error) {
	return s.getFriendById(id), nil
}

func (s *FriendService) AddFriend(image string, name string) (*entity.Friend, error) {
	dbSize := int64(len(s.DB))
	s.DB = append(s.DB, &entity.Friend{
		ID:     dbSize + 1,
		Name:   name,
		Image:  image,
		Active: true,
	})
	return s.DB[dbSize], nil
}

func (s *FriendService) UpdateFriend(id int64, image string, name string) (*entity.Friend, error) {
	fr := s.getFriendById(id)

	if fr == nil {
		return nil, nil
	}

	fr.Name = name
	fr.Image = image
	return fr, nil
}

func (s *FriendService) DeleteFriend(id int64) (int, error) {
	fr := s.getFriendById(id)

	if fr == nil {
		return 0, nil
	}

	fr.Active = false

	return 1, nil
}

func (s *FriendService) getFriendById(id int64) *entity.Friend {
	for _, fr := range s.DB {
		if id == fr.ID && fr.Active {
			return fr
		}
	}

	return nil
}
