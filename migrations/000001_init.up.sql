CREATE TABLE hospitals (
    code        text PRIMARY KEY,
    name        text NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE staff (
    id             uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hospital_code  text NOT NULL REFERENCES hospitals(code),
    username       text NOT NULL,
    password_hash  text NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (hospital_code, username)
);

CREATE TABLE patients (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    hospital_code   text NOT NULL REFERENCES hospitals(code),
    patient_hn      text NOT NULL,
    national_id     text,
    passport_id     text,
    first_name_th   text,
    middle_name_th  text,
    last_name_th    text,
    first_name_en   text,
    middle_name_en  text,
    last_name_en    text,
    date_of_birth   date,
    phone_number    text,
    email           text,
    gender          char(1) CHECK (gender IN ('M', 'F')),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (hospital_code, patient_hn)
);

CREATE INDEX patients_hospital_national_id_idx ON patients (hospital_code, national_id);
CREATE INDEX patients_hospital_passport_id_idx ON patients (hospital_code, passport_id);

INSERT INTO hospitals (code, name) VALUES
    ('hospital-a', 'Hospital A'),
    ('hospital-b', 'Hospital B');
