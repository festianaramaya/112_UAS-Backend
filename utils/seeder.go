package utils

import (
	"database/sql"
	"log"
    // Import model dan lainnya jika diperlukan
)

// Pastikan Anda juga memiliki fungsi HashPassword() di paket utils.

// SetupDatabase menjalankan DDL, Seeding Roles & Permissions, dan membuat User Admin pertama.
func SetupDatabase(db *sql.DB) error {
	log.Println("Running DDL and Seeding...")

	// 1. Jalankan DDL (Membuat Tabel dan Tipe ENUM)
	if err := runDDL(db); err != nil {
		return err
	}

	// 2. Seeding Roles dan Permissions
	if err := seedRolesAndPermissions(db); err != nil {
		return err
	}

	// 3. Seeding Admin User (FR-009)
	if err := seedAdminUser(db); err != nil {
		return err
	}

	return nil
}

// runDDL menjalankan semua query CREATE TABLE dan CREATE TYPE.
func runDDL(db *sql.DB) error {
	// DDL Queries (Diambil dari skema SRS)
	queries := []string{
		// 1. Wajib: UUID Extension
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`,

		// 2. Tipe ENUM untuk Achievement Status (FIX: Menggunakan DO $$ untuk menghindari syntax error "at or near NOT")
		`DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'achievement_status') THEN
				CREATE TYPE achievement_status AS ENUM (
					'draft', 
					'submitted', 
					'verified', 
					'rejected'
				);
			END IF;
		END
		$$ LANGUAGE plpgsql;`,

		// 3. Tabel roles
		`CREATE TABLE IF NOT EXISTS roles (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name VARCHAR(50) UNIQUE NOT NULL,
            description TEXT,
            created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
        );`,

		// 4. Tabel users
		`CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            username VARCHAR(50) UNIQUE NOT NULL,
            email VARCHAR(100) UNIQUE NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            full_name VARCHAR(100) NOT NULL,
            role_id UUID NOT NULL REFERENCES roles(id),
            is_active BOOLEAN NOT NULL DEFAULT true,
            created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
        );`,

		// 5. Tabel permissions
		`CREATE TABLE IF NOT EXISTS permissions (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            name VARCHAR(100) UNIQUE NOT NULL,
            resource VARCHAR(50) NOT NULL,
            action VARCHAR(50) NOT NULL,
            description TEXT
        );`,

		// 6. Tabel role_permissions
		`CREATE TABLE IF NOT EXISTS role_permissions (
            role_id UUID NOT NULL REFERENCES roles(id),
            permission_id UUID NOT NULL REFERENCES permissions(id),
            PRIMARY KEY (role_id, permission_id)
        );`,

		// 7. Tabel lecturers
		`CREATE TABLE IF NOT EXISTS lecturers (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL UNIQUE REFERENCES users(id),
            lecturer_id VARCHAR(20) UNIQUE NOT NULL,
            department VARCHAR(100),
            created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
        );`,

		// 8. Tabel students
		`CREATE TABLE IF NOT EXISTS students (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            user_id UUID NOT NULL UNIQUE REFERENCES users(id),
            student_id VARCHAR(20) UNIQUE NOT NULL,
            program_study VARCHAR(100),
            academic_year VARCHAR(10),
            advisor_id UUID REFERENCES lecturers(id),
            created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
        );`,

		// 9. Tabel achievement_references (Menggunakan tipe achievement_status yang dibuat di atas)
		`CREATE TABLE IF NOT EXISTS achievement_references (
            id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
            student_id UUID NOT NULL REFERENCES students(id),
            mongo_achievement_id VARCHAR(24) NOT NULL,
            status achievement_status NOT NULL, 
            submitted_at TIMESTAMP WITHOUT TIME ZONE,
            verified_at TIMESTAMP WITHOUT TIME ZONE,
            verified_by UUID REFERENCES users(id),
            rejection_note TEXT,
            created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW(),
            updated_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT NOW()
        );`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Printf("ERROR DDL: %v", err)
			return err
		}
	}
	log.Println("Database DDL completed.")
	return nil
}

// seedRolesAndPermissions (Tetap sama)
// ... (Logika seeding roles, permissions, dan relasi)

// seedAdminUser (Tetap sama)
// ... (Logika pembuatan user admin)


// ********** Logika seedRolesAndPermissions dan seedAdminUser di bawah ini tetap dipertahankan **********

func seedRolesAndPermissions(db *sql.DB) error {
	roles := map[string]string{
		"Admin":       "Pengelola sistem dengan hak akses penuh",
		"Dosen Wali":  "Verifikator prestasi mahasiswa bimbingan",
		"Mahasiswa":   "Pelapor prestasi",
	}

	// 1. Insert Roles
	for name, desc := range roles {
		var roleID string
		err := db.QueryRow("SELECT id FROM roles WHERE name = $1", name).Scan(&roleID)
		if err == sql.ErrNoRows {
			err = db.QueryRow(`INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id`, name, desc).Scan(&roleID)
			if err != nil {
				return err
			}
			log.Printf("Role '%s' added.", name)
		} else if err != nil {
			return err
		}
	}

	// 2. Insert Permissions
	permissionsData := []struct{ Name, Resource, Action, Description string }{
		{"user:manage", "user", "manage", "CRUD dan assign role user"},
		{"achievement:create", "achievement", "create", "Membuat prestasi draft"},
		{"achievement:read", "achievement", "read", "Melihat daftar dan detail prestasi"},
		{"achievement:update", "achievement", "update", "Memperbarui prestasi draft"},
		{"achievement:delete", "achievement", "delete", "Menghapus prestasi draft"},
		{"achievement:verify", "achievement", "verify", "Verifikasi prestasi yang submitted"},
	}

	for _, p := range permissionsData {
		var permID string
		err := db.QueryRow("SELECT id FROM permissions WHERE name = $1", p.Name).Scan(&permID)
		if err == sql.ErrNoRows {
			err = db.QueryRow(`INSERT INTO permissions (name, resource, action, description) VALUES ($1, $2, $3, $4) RETURNING id`, p.Name, p.Resource, p.Action, p.Description).Scan(&permID)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}

	// 3. Assign Permissions
	assignPermissions := func(roleName string, perms []string) error {
		var roleID string
		if err := db.QueryRow("SELECT id FROM roles WHERE name = $1", roleName).Scan(&roleID); err != nil {
			return err
		}

		for _, permName := range perms {
			var permID string
			if err := db.QueryRow("SELECT id FROM permissions WHERE name = $1", permName).Scan(&permID); err != nil {
				return err
			}
			_, err := db.Exec(`INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, roleID, permID)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if err := assignPermissions("Admin", []string{"user:manage", "achievement:create", "achievement:read", "achievement:update", "achievement:delete", "achievement:verify"}); err != nil {
		return err
	}
	if err := assignPermissions("Dosen Wali", []string{"achievement:read", "achievement:verify"}); err != nil {
		return err
	}
	if err := assignPermissions("Mahasiswa", []string{"achievement:create", "achievement:read", "achievement:update", "achievement:delete"}); err != nil {
		return err
	}

	log.Println("Roles and Permissions seeded.")
	return nil
}

func seedAdminUser(db *sql.DB) error {
	const defaultUsername = "admin"
	const defaultPassword = "12345678"

	var userID string
	// Cek apakah user admin sudah ada
	err := db.QueryRow("SELECT id FROM users WHERE username = $1", defaultUsername).Scan(&userID)
	if err == nil {
		return nil // User sudah ada
	}
	if err != sql.ErrNoRows {
		return err
	}

	// 1. Dapatkan Role ID
	var adminRoleID string
	if err := db.QueryRow("SELECT id FROM roles WHERE name = 'Admin'").Scan(&adminRoleID); err != nil {
		return err
	}

	// 2. Hash Password
	hashedPassword, err := HashPassword(defaultPassword)
	if err != nil {
		return err
	}

	// 3. Insert User
	query := `
        INSERT INTO users (username, email, password_hash, full_name, role_id)
        VALUES ($1, $2, $3, $4, $5) 
        RETURNING id
    `
	err = db.QueryRow(query,
		defaultUsername,
		"admin@gmail.com",
		hashedPassword,
		"System Administrator",
		adminRoleID,
	).Scan(&userID)

	if err != nil {
		return err
	}

	log.Printf("User Admin created (username: %s, pass: %s!)", defaultUsername, defaultPassword)
	return nil
}