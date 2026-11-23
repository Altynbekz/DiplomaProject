package httpapi

import (
	"net/http"

	"diploma/internal/service"
)

func (a *API) createVerification(w http.ResponseWriter, r *http.Request) error {
	findingID, err := id(r)
	if err != nil {
		return err
	}
	var input service.VerificationInput
	if err := body(w, r, &input); err != nil {
		return err
	}
	review, err := a.Service.CreateVerification(r.Context(), findingID, input, requestID(r))
	if err != nil {
		return err
	}
	write(w, http.StatusCreated, review)
	return nil
}

func (a *API) verifications(w http.ResponseWriter, r *http.Request) error {
	findingID, err := id(r)
	if err != nil {
		return err
	}
	limit, cursor, err := pagination(r)
	if err != nil {
		return err
	}
	page, err := a.Service.Verifications(r.Context(), findingID, limit, cursor)
	if err != nil {
		return err
	}
	write(w, http.StatusOK, page)
	return nil
}
