package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Nutto555/hospital-middleware/internal/domain"
)

// PatientService is what the search endpoint needs from the service layer.
type PatientService interface {
	Search(ctx context.Context, hospitalCode string, c domain.SearchCriteria) ([]domain.Patient, error)
}

const dateLayout = "2006-01-02"

type searchRequest struct {
	NationalID  *string `json:"national_id"`
	PassportID  *string `json:"passport_id"`
	FirstName   *string `json:"first_name"`
	MiddleName  *string `json:"middle_name"`
	LastName    *string `json:"last_name"`
	DateOfBirth *string `json:"date_of_birth"`
	PhoneNumber *string `json:"phone_number"`
	Email       *string `json:"email"`
}

// criteria trims every field, drops blank ones and parses the date.
func (r searchRequest) criteria() (domain.SearchCriteria, error) {
	c := domain.SearchCriteria{
		NationalID:  clean(r.NationalID),
		PassportID:  clean(r.PassportID),
		FirstName:   clean(r.FirstName),
		MiddleName:  clean(r.MiddleName),
		LastName:    clean(r.LastName),
		PhoneNumber: clean(r.PhoneNumber),
		Email:       clean(r.Email),
	}
	if c.Email != nil {
		addr, err := mail.ParseAddress(*c.Email)
		if err != nil {
			return domain.SearchCriteria{}, errors.New("email is not a valid address")
		}
		c.Email = &addr.Address
	}
	if d := clean(r.DateOfBirth); d != nil {
		t, err := time.Parse(dateLayout, *d)
		if err != nil {
			return domain.SearchCriteria{}, errors.New("date_of_birth must be YYYY-MM-DD")
		}
		c.DateOfBirth = &t
	}
	return c, nil
}

func clean(s *string) *string {
	if s == nil {
		return nil
	}
	t := strings.TrimSpace(*s)
	if t == "" {
		return nil
	}
	return &t
}

type patientResponse struct {
	ID           string  `json:"id"`
	Hospital     string  `json:"hospital"`
	PatientHN    string  `json:"patient_hn"`
	NationalID   *string `json:"national_id"`
	PassportID   *string `json:"passport_id"`
	FirstNameTH  *string `json:"first_name_th"`
	MiddleNameTH *string `json:"middle_name_th"`
	LastNameTH   *string `json:"last_name_th"`
	FirstNameEN  *string `json:"first_name_en"`
	MiddleNameEN *string `json:"middle_name_en"`
	LastNameEN   *string `json:"last_name_en"`
	DateOfBirth  *string `json:"date_of_birth"`
	PhoneNumber  *string `json:"phone_number"`
	Email        *string `json:"email"`
	Gender       *string `json:"gender"`
}

type searchResponse struct {
	Patients []patientResponse `json:"patients"`
}

func toPatientResponse(p domain.Patient) patientResponse {
	out := patientResponse{
		ID: p.ID, Hospital: p.HospitalCode, PatientHN: p.PatientHN,
		NationalID: p.NationalID, PassportID: p.PassportID,
		FirstNameTH: p.FirstNameTH, MiddleNameTH: p.MiddleNameTH, LastNameTH: p.LastNameTH,
		FirstNameEN: p.FirstNameEN, MiddleNameEN: p.MiddleNameEN, LastNameEN: p.LastNameEN,
		PhoneNumber: p.PhoneNumber, Email: p.Email, Gender: p.Gender,
	}
	if p.DateOfBirth != nil {
		s := p.DateOfBirth.Format(dateLayout)
		out.DateOfBirth = &s
	}
	return out
}

func searchPatients(svc PatientService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var in searchRequest
		if err := c.ShouldBindJSON(&in); err != nil && !errors.Is(err, io.EOF) {
			writeError(c, http.StatusBadRequest, "invalid request body")
			return
		}
		criteria, err := in.criteria()
		if err != nil {
			writeError(c, http.StatusBadRequest, err.Error())
			return
		}
		principal, ok := principalFrom(c)
		if !ok {
			writeError(c, http.StatusUnauthorized, "invalid or missing token")
			return
		}
		patients, err := svc.Search(c.Request.Context(), principal.HospitalCode, criteria)
		if err != nil {
			writeServiceError(c, err)
			return
		}
		out := searchResponse{Patients: make([]patientResponse, 0, len(patients))}
		for _, p := range patients {
			out.Patients = append(out.Patients, toPatientResponse(p))
		}
		c.JSON(http.StatusOK, out)
	}
}
