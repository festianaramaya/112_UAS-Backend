package repository

import (
	"context"
	"errors"
	"uas/app/model"
	"github.com/jmoiron/sqlx" 
)

// GANTI: struct DB menggunakan *sqlx.DB
type UserRepository struct {
	DB *sqlx.DB
}

// GANTI: Constructor menerima *sqlx.DB
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) GetUserPermissions(roleID string) ([]string, error) {
    // Note: Parameter roleID harus string karena di model Go Anda RoleID adalah string.

	query := `
		SELECT p.name
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = $1
	`
	rows, err := r.DB.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var permName string
		if err := rows.Scan(&permName); err != nil {
			return nil, err
		}
		permissions = append(permissions, permName)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return permissions, nil
}


// Tambahkan GetRoleNameByID untuk melengkapi data JWT
func (r *UserRepository) GetRoleNameByID(roleID string) (string, error) {
    var roleName string
    query := `SELECT name FROM roles WHERE id = $1 LIMIT 1`
    
    err := r.DB.QueryRow(query, roleID).Scan(&roleName)
    if err != nil {
        return "", err
    }
    return roleName, nil
}

func (r *UserRepository) GetAll() ([]model.User, error) {
	// FIX: Tambahkan password_hash untuk konsistensi, meskipun biasanya tidak dikirim ke client
	rows, err := r.DB.Query(`
		SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at, password_hash
		FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
        var passwordHash string // Placeholder sementara
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.FullName,
			&u.RoleID,
			&u.IsActive,
			&u.CreatedAt,
			&u.UpdatedAt,
            &passwordHash, // Tangkap hash
		); err != nil {
			return nil, err
		}
        u.PasswordHash = passwordHash // Set kembali
		users = append(users, u)
	}

	return users, nil
}

func (r *UserRepository) GetUserByID(id string) (*model.User, error) {
    // FIX: Menerima ID bertipe string (UUID)
	var u model.User
    var passwordHash string
	query := `
		SELECT id, username, email, full_name, role_id, is_active, created_at, updated_at, password_hash
		FROM users WHERE id=$1`

    // FIX: Gunakan id bertipe string
	err := r.DB.QueryRow(query, id).Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.FullName,
		&u.RoleID,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
        &passwordHash,
	)
    u.PasswordHash = passwordHash

	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*model.User, error) {
    // Note: Query ini sudah benar dan mencakup password_hash
	query := `
		SELECT id, username, email, password_hash, full_name, role_id, is_active, created_at, updated_at
		FROM users
		WHERE username = $1
		LIMIT 1
	`
    // ... (Logika FindByUsername sudah benar)
	row := r.DB.QueryRowContext(ctx, query, username)

	var u model.User
	err := row.Scan(
		&u.ID,
		&u.Username,
		&u.Email,
		&u.PasswordHash,
		&u.FullName,
		&u.RoleID,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}

// Tambahkan di bawah FindByUsername
func (r *UserRepository) Create(u *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
	`

	row := r.DB.QueryRow(query,
		u.Username,
		u.Email,
		u.PasswordHash,
		u.FullName,
		u.RoleID,
		u.IsActive,
	)
    // Tangkap ID dan timestamps yang dihasilkan DB
	return row.Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

// app/repository/user_repository.go

// UpdateRole memperbarui role_id dari user spesifik
func (r *UserRepository) UpdateRole(userID, roleID string) error {
    query := `
        UPDATE users
        SET role_id = $1, updated_at = NOW()
        WHERE id = $2
    `
    result, err := r.DB.Exec(query, roleID, userID)
    if err != nil {
        return err
    }
    
    // Opsional: Cek apakah ada baris yang terpengaruh (terupdate)
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        // Jika 0, ID user tidak ditemukan
        return errors.New("user not found or role already set") 
    }
    return nil
}

// SoftDelete menandai user sebagai tidak aktif
func (r *UserRepository) SoftDelete(userID string) error {
    query := `
        UPDATE users
        SET is_active = FALSE, updated_at = NOW()
        WHERE id = $1 AND is_active = TRUE
    `
    result, err := r.DB.Exec(query, userID)
    if err != nil {
        return err
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return err
    }
    if rowsAffected == 0 {
        return errors.New("user not found or already deleted") 
    }
    return nil
}