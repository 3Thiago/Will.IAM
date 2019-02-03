package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ghostec/Will.IAM/models"
	"github.com/ghostec/Will.IAM/usecases"
	"github.com/topfreegames/extensions/middleware"
)

func servicesListHandler(
	ssUC usecases.Services,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		ssSl, err := ssUC.WithContext(r.Context()).List()
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bts, err := keepJSONFieldsBytes(ssSl, "id", "name", "created_at", "updated_at")
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteBytes(w, 200, bts)
	}
}

func servicesCreateHandler(
	ssUC usecases.Services,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			l.WithError(err).Error("servicesCreateHandler ioutil.ReadAll failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		service := &models.Service{}
		err = json.Unmarshal(body, service)
		if err != nil {
			l.WithError(err).Error("servicesCreateHandler json.Unmarshal failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		v := service.Validate()
		if !v.Valid() {
			WriteBytes(w, http.StatusUnprocessableEntity, v.Errors())
			return
		}
		saID, _ := getServiceAccountID(r.Context())
		service.CreatorServiceAccountID = saID
		if err := ssUC.WithContext(r.Context()).Create(service); err != nil {
			l.WithError(err).Error("servicesCreateHandler ssUC.Create failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
