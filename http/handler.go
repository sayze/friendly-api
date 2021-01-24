/** File contains the route handlers/controllers for the server package. */

package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// createFriendRequest defines a request for creating a Friend entity
type createFriendRequest struct {
	Name  string `json:"name" validate:"required,max=20,min=2"`
	Image string `json:"image"`
}

// updateFriendRequest defines a request for updating a Friend entity
type updateFriendRequest struct {
	ID    int64  `json:"id" validate:"required,numeric"`
	Name  string `json:"name" validate:"required,max=20,min=2"`
	Image string `json:"image"`
}

func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintln(w, `{"status":"ok"}`)

	if err != nil {
		render.Render(w, r, ErrFatalRequest(err))
		return
	}
}

func (h *Handler) HandleCreateFriend(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if total := h.FriendService.CountFriends(); total > 100 {
		render.Render(w, r, &ErrResponse{
			HTTPStatusCode: http.StatusBadRequest,
			ErrorText:      "maximum friends exceeded",
			StatusText:     "server error",
		})
		return
	}

	err = validate.Struct(createFriendRequest{
		Name: r.FormValue("name"),
	})

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	friend, err := h.FriendService.AddFriend("", r.FormValue("name"))

	if err != nil {
		render.Render(w, r, ErrFatalRequest(err))
		return
	}

	// handle image upload
	file, handler, err := r.FormFile("image")

	switch err {
	case nil:
		defer file.Close()
		fid := strconv.FormatInt(friend.ID, 10)
		image, err := h.Cdn.uploadImage(file, handler.Filename, fid)

		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		if _, err = h.FriendService.UpdateFriend(friend.ID, image, ""); err != nil {
			render.Render(w, r, SuccessNoContentRequest("Could not find friend with id "+fid))
		}

		break
	case http.ErrMissingFile:
		break
	default:
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, SuccessDataRequest(friend))
}

func (h *Handler) HandleGetFriend(w http.ResponseWriter, r *http.Request) {
	fid := chi.URLParam(r, "id")
	search := r.URL.Query().Get("search")

	if len(fid) == 0 {
		friends, err := h.FriendService.All(search)

		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}

		render.Render(w, r, SuccessDataRequest(friends))
		return
	}

	err := validate.Var(fid, "numeric")

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid friend id provided")))
		return
	}

	friendId, _ := strconv.ParseInt(fid, 10, 64)

	if friend, _ := h.FriendService.GetFriend(friendId); friend == nil {
		render.Render(w, r, SuccessNoContentRequest("Could not find friend with id "+fid))
	} else {
		render.Render(w, r, SuccessDataRequest(friend))
	}
}

func (h *Handler) HandleDeleteFriend(w http.ResponseWriter, r *http.Request) {
	fid := chi.URLParam(r, "id")

	err := validate.Var(fid, "numeric")

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(errors.New("invalid friend id provided")))
		return
	}

	friendId, _ := strconv.ParseInt(fid, 10, 64)

	err = h.Cdn.deleteImage(fid)

	if err != nil {
		render.Render(w, r, ErrFatalRequest(err))
		return
	}

	if ct, _ := h.FriendService.DeleteFriend(friendId); ct < 1 {
		render.Render(w, r, SuccessNoContentRequest("Could not find friend with id "+fid))
	} else {
		render.Render(w, r, SuccessDataRequest("Friend removed successfully"))
	}
}

func (h *Handler) HandleUpdateFriend(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)

	if err != nil {
		render.Render(w, r, ErrFatalRequest(err))
		return
	}

	fr := updateFriendRequest{
		ID:   id,
		Name: r.FormValue("name"),
	}

	err = validate.Struct(fr)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	file, handler, err := r.FormFile("image")

	var newImage string

	switch err {
	case nil:
		defer file.Close()
		newImage, err = h.Cdn.uploadImage(file, handler.Filename, strconv.FormatInt(id, 10))

		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		break
	case http.ErrMissingFile:
		break
	default:
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if friend, _ := h.FriendService.UpdateFriend(fr.ID, newImage, fr.Name); friend == nil {
		fidStr := strconv.FormatInt(fr.ID, 10)
		render.Render(w, r, SuccessNoContentRequest("Could not find friend with id "+fidStr))
	} else {
		render.Render(w, r, SuccessDataRequest(friend))
	}
}

func slugify(in string) string {
	return strings.ToLower(strings.ReplaceAll(in, " ", "-"))
}
