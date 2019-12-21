package memory

import (
	"github.com/sayze/foundu-taker-api/internal/entity"
)

func (c *Client) getFriendById(id int64) *entity.Friend {
	for _, friend := range c.friends {
		if id == friend.ID && friend.Active{
			return friend
		}
	}

	return nil
}

func (c *Client) CreateFriend(friend entity.Friend) (*entity.Friend, error) {
	friendLength := int64(len(c.friends))
	friend.ID = friendLength + 1
	c.friends = append(c.friends, &friend)
	return &friend, nil
}

func (c *Client) ViewFriend(id int64) (*entity.Friend, error) {
	return c.getFriendById(id), nil
}

func (c *Client) DeleteFriend(id int64) (bool, error) {
	friend := c.getFriendById(id)

	if friend == nil {
		return false, nil
	}

	friend.Active = false

	return true, nil
}

func (c *Client) UpdateFriend(friend *entity.FriendUpdate) (*entity.Friend, error) {
	fr := c.getFriendById(friend.ID)

	if fr == nil {
		return nil, nil
	}

	fr.Name = friend.Name
	fr.Age = friend.Age

	return fr, nil
}
