package httpapi

import (
	"net/http"

	"diploma/internal/domain"
	"diploma/internal/service"
)

func (a *API) createAssessment(w http.ResponseWriter, r *http.Request) error {
	findingID, err := id(r)
	if err != nil {
		return err
	}
	var input service.AssessmentInput
	if err := body(w, r, &input); err != nil {
		return err
	}
	snapshot, err := a.Service.CreateAssessment(r.Context(), findingID, input, requestID(r))
	if err != nil {
		return err
	}
	w.Header().Set("Location", "/api/v1/assessments/"+snapshot.ID)
	write(w, http.StatusCreated, snapshot)
	return nil
}

func (a *API) assessments(w http.ResponseWriter, r *http.Request) error {
	findingID, err := id(r)
	if err != nil {
		return err
	}
	limit, cursor, err := pagination(r)
	if err != nil {
		return err
	}
	result, err := a.Service.Assessments(r.Context(), findingID, limit, cursor)
	if err != nil {
		return err
	}
	write(w, http.StatusOK, result)
	return nil
}

func (a *API) assessment(w http.ResponseWriter, r *http.Request) error {
	assessmentID, err := id(r)
	if err != nil {
		return err
	}
	if r.URL.RawQuery != "" {
		return domain.Fail(400, "invalid_request", "", "Assessment detail takes no query parameters")
	}
	result, err := a.Service.GetAssessment(r.Context(), assessmentID)
	if err != nil {
		return err
	}
	write(w, http.StatusOK, result)
	return nil
}
