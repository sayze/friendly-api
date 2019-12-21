/**

File contains the route handlers/controllers for the server package.

HandleCreateFriend - POST /friend
------------------
Creates a new friend with the data supplied in request body

HandleDeleteFriend - DELETE /friend/{id}
------------------
Performs soft delete against the supplied id (if valid)

HandleViewFriend - GET /friend/{id}
------------------
Retrieves a single friend based on supplied id (if valid)

HandleUpdateFriend - PATCH /friend
------------------
Performs an update to a friend given the request body contains a valid id

Improvements

- Group logic to check existence of friend as one function rather than repeating.

 */

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

// updateFriendRequest defines a request for updating a Friend entity
type updateFriendRequest struct {
	ID   int64  `json:"id" validate:"required,numeric"`
	Name string `json:"name" validate:"required,max=20,min=2"`
	Age  int    `json:"age" validate:"required,numeric"`
}

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

func (s *Server) HandleUpdateFriend(w http.ResponseWriter, r *http.Request) {

	fr := &updateFriendRequest{}

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

	if friend, _ := s.friendStore.UpdateFriend(&entity.FriendUpdate{ID: fr.ID, Name: fr.Name, Age: fr.Age}); friend == nil {
		fidStr := strconv.FormatInt(fr.ID, 10)
		render.JSON(w, r, "Could not find friend with id " + fidStr)
	} else {
		render.JSON(w, r, friend)
	}
}
