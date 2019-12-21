package server

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/sayze/foundu-taker-api/internal/entity"
	"net/http"
	"strconv"
)

// createFriendRequest defines a request for creating a Friend entity
type createFriendRequest struct {
	Name string `json:"name" validate:"required,max=20,min=2"`
	Age  int    `json:"age" validate:"required,numeric"`
}

// HandleCreateFriend is controller for POST /friend
func (s *Server) HandleCreateFriend(w http.ResponseWriter, r *http.Request) {
	fr := &createFriendRequest{}

	err := render.DecodeJSON(r.Body, fr)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	// Validate the struct against rules.
	err = validate.Struct(fr)

	if err != nil {
		// handle error
		return
	}

	friend, _ := s.friendStore.CreateFriend(entity.Friend{Name: fr.Name, Age: fr.Age, Active: true})

	render.JSON(w, r, friend)
}

// HandleGetFriend is controller for Get /friend
func (s *Server) HandleGetFriend(w http.ResponseWriter, r *http.Request) {
	fid := chi.URLParam(r, "id")

	err := validate.Var(fid, "numeric")

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid friend id provided")))
		return
	}

	friendId, _ := strconv.ParseInt(fid, 10, 64)

	friend, _ := s.friendStore.ViewFriend(friendId)

	if friend == nil {
		render.JSON(w, r, "Could not find friend with id "+fid)
	} else {
		render.JSON(w, r, friend)
	}
}

func (s *Server) HandleDeleteFriend(w http.ResponseWriter, r *http.Request) {
	fid := chi.URLParam(r, "id")

	err := validate.Var(fid, "numeric")

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid friend id provided")))
		return
	}

	friendId, _ := strconv.ParseInt(fid, 10, 64)

	if ok, _ := s.friendStore.DeleteFriend(friendId); !ok {
		render.JSON(w, r, "Could not find friend with id "+fid)
	} else {
		render.JSON(w, r, "Successfully deleted friend")
	}
}
