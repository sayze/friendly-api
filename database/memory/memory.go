// Package DB implements an in DB storage interface for all the entities.
package memory

import (
	"strings"

	"github.com/sayze/friendly-api/entity"
)

type FriendService struct {
	DB []*entity.Friend
}

func NewService() *FriendService {
	return &FriendService{}
}

func (s *FriendService) All(search string) ([]*entity.Friend, error) {
	var activeFriends []*entity.Friend
	searchToLower := strings.ToLower(search)

	for _, fr := range s.DB {
		nameToLower := strings.ToLower(fr.Name)

		if len(searchToLower) == 0 || strings.Contains(nameToLower, searchToLower) {
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
		ID:    dbSize + 1,
		Name:  name,
		Image: image,
	})
	return s.DB[dbSize], nil
}

func (s *FriendService) UpdateFriend(id int64, image string, name string) (*entity.Friend, error) {
	fr := s.getFriendById(id)

	if fr == nil {
		return nil, nil
	}

	if len(name) != 0 {
		fr.Name = name
	}

	if len(image) != 0 {
		fr.Image = image
	}

	return fr, nil
}

func (s *FriendService) DeleteFriend(id int64) (int, error) {
	originalLen := len(s.DB)

	for idx, fr := range s.DB {
		if id == fr.ID {
			s.DB = append(s.DB[:idx], s.DB[idx+1:]...)
		}
	}

	return originalLen - len(s.DB), nil
}

func (s *FriendService) getFriendById(id int64) *entity.Friend {
	for _, fr := range s.DB {
		if id == fr.ID {
			return fr
		}
	}

	return nil
}

func (s *FriendService) CountFriends() int {
	return len(s.DB)
}
