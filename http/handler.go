/** File contains the route handlers/controllers for the server package. */

package http

import (
	"errors"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"mime/multipart"
	"net/http"
	"strconv"
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

func (h *Handler) HandleCreateFriend(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	file, err := uploadImageFromRequest(r)

	if err != nil && err != http.ErrMissingFile {
		render.Render(w, r, ErrFatalRequest(err))
		return
	}

	// Call cdn for file upload.
	h.Cdn.UploadImage(file.)

	err = validate.Struct(createFriendRequest{
		Name:  r.FormValue("name"),
		Image: filename,
	})

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	friend, err := h.FriendService.AddFriend(filename, r.FormValue("name"))

	if err != nil {
		render.Render(w, r, ErrFatalRequest(err))
		return
	}

	render.Render(w, r, SuccessDataRequest(friend))
}

func (h *Handler) HandleGetFriend(w http.ResponseWriter, r *http.Request) {
	fid := chi.URLParam(r, "id")
	search := r.URL.Query().Get("search")

	if len(fid) == 0 {
		// TODO: Should probably have a search functioned as service definition.
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

	filename, err := uploadImageFromRequest(r)

	if err != nil && err != http.ErrMissingFile {
		render.Render(w, r, ErrFatalRequest(err))
		return
	}

	id, err := strconv.ParseInt(r.FormValue("id"), 10, 64)

	if err != nil {
		render.Render(w, r, ErrFatalRequest(err))
		return
	}

	fr := updateFriendRequest{
		ID:    id,
		Name:  r.FormValue("name"),
		Image: filename,
	}

	err = validate.Struct(fr)

	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if friend, _ := h.FriendService.UpdateFriend(fr.ID, "", fr.Name); friend == nil {
		fidStr := strconv.FormatInt(fr.ID, 10)
		render.Render(w, r, SuccessNoContentRequest("Could not find friend with id "+fidStr))
	} else {
		render.Render(w, r, SuccessDataRequest(friend))
	}
}

func uploadImageFromRequest(r *http.Request) (multipart.File, error) {




	//f, err := os.OpenFile("./public/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	//
	//if err != nil {
	//	return "", err
	//}
	//
	//defer f.Close()
	//
	//_, err = io.Copy(f, file)
	//
	//if err != nil {
	//	return "", err
	//}

	return file, nil
}
