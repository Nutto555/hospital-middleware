package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nutto555/hospital-middleware/internal/domain"
)

// searchLimit caps a search result; the API has no pagination.
const searchLimit = 100

const patientColumns = `id::text, hospital_code, patient_hn, national_id, passport_id,
	first_name_th, middle_name_th, last_name_th, first_name_en, middle_name_en, last_name_en,
	date_of_birth, phone_number, email, gender`

// PatientRepo stores the local copy of hospital patient records.
type PatientRepo struct {
	db *pgxpool.Pool
}

func NewPatientRepo(db *pgxpool.Pool) *PatientRepo {
	return &PatientRepo{db: db}
}

// Search returns the hospital's patients matching every provided criterion, ordered by
// hospital number and capped at searchLimit. Names match the Thai or English column.
func (r *PatientRepo) Search(ctx context.Context, hospitalCode string, c domain.SearchCriteria) ([]domain.Patient, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+patientColumns+`
		 FROM patients
		 WHERE hospital_code = $1
		   AND ($2::text IS NULL OR national_id = $2)
		   AND ($3::text IS NULL OR passport_id = $3)
		   AND ($4::text IS NULL OR lower(first_name_th) = lower($4) OR lower(first_name_en) = lower($4))
		   AND ($5::text IS NULL OR lower(middle_name_th) = lower($5) OR lower(middle_name_en) = lower($5))
		   AND ($6::text IS NULL OR lower(last_name_th) = lower($6) OR lower(last_name_en) = lower($6))
		   AND ($7::date IS NULL OR date_of_birth = $7)
		   AND ($8::text IS NULL OR phone_number = $8)
		   AND ($9::text IS NULL OR lower(email) = lower($9))
		 ORDER BY patient_hn
		 LIMIT `+fmt.Sprint(searchLimit),
		hospitalCode, c.NationalID, c.PassportID, c.FirstName, c.MiddleName, c.LastName,
		c.DateOfBirth, c.PhoneNumber, c.Email,
	)
	if err != nil {
		return nil, fmt.Errorf("search patients: %w", err)
	}
	defer rows.Close()

	var patients []domain.Patient
	for rows.Next() {
		var p domain.Patient
		if err := rows.Scan(&p.ID, &p.HospitalCode, &p.PatientHN, &p.NationalID, &p.PassportID,
			&p.FirstNameTH, &p.MiddleNameTH, &p.LastNameTH, &p.FirstNameEN, &p.MiddleNameEN, &p.LastNameEN,
			&p.DateOfBirth, &p.PhoneNumber, &p.Email, &p.Gender); err != nil {
			return nil, fmt.Errorf("scan patient: %w", err)
		}
		patients = append(patients, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search patients: %w", err)
	}
	return patients, nil
}

// Upsert inserts the record or, when the hospital already has that hospital number, replaces
// every field with the newer values. The id of an existing row is kept.
func (r *PatientRepo) Upsert(ctx context.Context, p domain.Patient) (domain.Patient, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO patients (hospital_code, patient_hn, national_id, passport_id,
		     first_name_th, middle_name_th, last_name_th, first_name_en, middle_name_en, last_name_en,
		     date_of_birth, phone_number, email, gender)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		 ON CONFLICT (hospital_code, patient_hn) DO UPDATE SET
		     national_id = EXCLUDED.national_id, passport_id = EXCLUDED.passport_id,
		     first_name_th = EXCLUDED.first_name_th, middle_name_th = EXCLUDED.middle_name_th,
		     last_name_th = EXCLUDED.last_name_th, first_name_en = EXCLUDED.first_name_en,
		     middle_name_en = EXCLUDED.middle_name_en, last_name_en = EXCLUDED.last_name_en,
		     date_of_birth = EXCLUDED.date_of_birth, phone_number = EXCLUDED.phone_number,
		     email = EXCLUDED.email, gender = EXCLUDED.gender, updated_at = now()
		 RETURNING id::text`,
		p.HospitalCode, p.PatientHN, p.NationalID, p.PassportID,
		p.FirstNameTH, p.MiddleNameTH, p.LastNameTH, p.FirstNameEN, p.MiddleNameEN, p.LastNameEN,
		p.DateOfBirth, p.PhoneNumber, p.Email, p.Gender,
	).Scan(&p.ID)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("upsert patient: %w", err)
	}
	return p, nil
}
