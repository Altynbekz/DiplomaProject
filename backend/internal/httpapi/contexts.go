package httpapi

import (
	"net/http"

	"diploma/internal/domain"
	"diploma/internal/service"
)

func (a *API) createContext(w http.ResponseWriter, r *http.Request) error {
	targetID, err := id(r)
	if err != nil {
		return err
	}
	var input service.ContextInput
	if err := body(w, r, &input); err != nil {
		return err
	}
	revision, err := a.Service.CreateContext(r.Context(), targetID, input, requestID(r))
	if err != nil {
		return err
	}
	write(w, http.StatusCreated, revision)
	return nil
}

func (a *API) contexts(w http.ResponseWriter, r *http.Request) error {
	targetID, err := id(r)
	if err != nil {
		return err
	}
	limit, cursor, err := pagination(r)
	if err != nil {
		return err
	}
	page, err := a.Service.Contexts(r.Context(), targetID, limit, cursor)
	if err != nil {
		return err
	}
	write(w, http.StatusOK, page)
	return nil
}

func (a *API) resolvedContext(w http.ResponseWriter, r *http.Request) error {
	findingID, err := id(r)
	if err != nil {
		return err
	}
	if r.URL.RawQuery != "" {
		return domain.Fail(400, "invalid_request", "", "Resolved context takes no query parameters")
	}
	result, err := a.Service.ResolvedContext(r.Context(), findingID)
	if err != nil {
		return err
	}
	write(w, http.StatusOK, result)
	return nil
}
