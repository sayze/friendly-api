package memory

import (
	"github.com/sayze/foundu-taker-api/internal/entity"
)

func (c *Client) CreateFriend(friend entity.Friend) (*entity.Friend, error) {
	friendLength := int64(len(c.friends))
	friend.ID = friendLength + 1
	c.friends = append(c.friends, &friend)

	return &friend, nil
}

func (c *Client) ViewFriend(id int64) (*entity.Friend, error) {
	for _, friend := range c.friends {
		if id == friend.ID && friend.Active{
			return friend, nil
		}
	}

	return nil, nil
}

func (c *Client) DeleteFriend(id int64) (bool, error) {
	for _, friend := range c.friends {
		if id == friend.ID && friend.Active{
			friend.Active = false
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) UpdateFriend(update *entity.FriendUpdate) (*entity.Friend, error) {
	// Update friend.
	return nil, nil
}
