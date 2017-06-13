package command

import (
	"encoding/json"
	"errors"
	"github.com/function61/pi-security-module/folder/event"
	"github.com/function61/pi-security-module/state"
	"github.com/function61/pi-security-module/util"
	"net/http"
)

type RenameFolderRequest struct {
	Id   string
	Name string
}

func (f *RenameFolderRequest) Validate() error {
	if f.Id == "" {
		return errors.New("Id missing")
	}
	if f.Name == "" {
		return errors.New("Name missing")
	}
	if state.FolderById(f.Id) == nil {
		return errors.New("Folder by Id not found")
	}

	return nil
}

func HandleRenameFolderRequest(w http.ResponseWriter, r *http.Request) {
	var req RenameFolderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	util.ApplyEvents([]interface{}{
		event.FolderRenamed{
			Id:   req.Id,
			Name: req.Name,
		},
	})

	w.Write([]byte("OK"))
}