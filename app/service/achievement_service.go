package service

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"uas/app/model"
	"uas/app/repository"
)

type AchievementService struct {
	PgRepo    *repository.AchievementRepository
	MongoRepo *repository.MongoAchievementRepository
}

func NewAchievementService(
	pg *repository.AchievementRepository,
	mongo *repository.MongoAchievementRepository,
) *AchievementService {
	return &AchievementService{
		PgRepo:    pg,
		MongoRepo: mongo,
	}
}

/* ===================== BASIC CRUD ===================== */

// GetAll godoc
// @Summary      Get all achievements
// @Description  Mengambil semua referensi prestasi dari PostgreSQL
// @Tags         Achievements
// @Produce      json
// @Success      200  {array}   model.AchievementReference
// @Security     BearerAuth
// @Router       /api/v1/achievements [get]
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	data, err := s.PgRepo.GetAll(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(data)
}

// GetDetail godoc
// @Summary      Get achievement detail
// @Description  Mengambil detail prestasi berdasarkan UUID PostgreSQL
// @Tags         Achievements
// @Param        id   path      string  true  "Achievement UUID"
// @Produce      json
// @Success      200  {object}  model.AchievementReference
// @Security     BearerAuth
// @Router       /api/v1/achievements/{id} [get]
func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
    id, err := uuid.Parse(c.Params("id"))
    if err != nil {
        return fiber.ErrBadRequest
    }

    // SEKARANG PANGGIL GetByID
    data, err := s.PgRepo.GetByID(c.Context(), id)
    if err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "Achievement not found"})
    }
    
    return c.JSON(data)
}

// Create godoc
// @Summary      Create achievement
// @Description  Membuat prestasi baru di MongoDB dan PostgreSQL (FR-003)
// @Tags         Achievements
// @Accept       json
// @Produce      json
// @Param        achievement  body      model.MongoAchievement  true  "Data Prestasi"
// @Success      201          {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /api/v1/achievements [post]
func (s *AchievementService) Create(c *fiber.Ctx) error {
	var mongoData model.MongoAchievement
	if err := c.BodyParser(&mongoData); err != nil {
		return fiber.ErrBadRequest
	}

	studentUUID, err := uuid.Parse(mongoData.StudentID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "student_id must be valid UUID")
	}

	ctx := c.Context()

	mongoID, err := s.MongoRepo.Create(ctx, &mongoData)
	if err != nil {
		return err
	}

	ref := model.AchievementReference{
		StudentID:          studentUUID.String(),
		MongoAchievementID: mongoID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.PgRepo.Create(&ref); err != nil {
		_ = s.MongoRepo.Delete(ctx, mongoID)
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "achievement created",
		"data":     ref,
	})
}

// Update godoc
// @Summary      Update achievement
// @Description  Memperbarui timestamp update prestasi di PostgreSQL
// @Tags         Achievements
// @Param        id   path      string  true  "Achievement UUID"
// @Produce      json
// @Success      200  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/achievements/{id} [put]
func (s *AchievementService) Update(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	if err := s.PgRepo.UpdateTimestamp(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{"message": "achievement updated"})
}

// Delete godoc
// @Summary      Delete achievement (Soft Delete)
// @Description  Mahasiswa menghapus prestasi draft dengan mengubah status menjadi 'deleted' (FR-005)
// @Tags         Achievements
// @Param        id   path      string  true  "Achievement UUID"
// @Produce      json
// @Success      200  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/achievements/{id} [delete]
func (s *AchievementService) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	// 1. Logika Soft Delete: Update status di PostgreSQL menjadi 'deleted'
	// Kita menggunakan UpdateStatus yang sudah ada, dengan status baru 'deleted'
	err = s.PgRepo.UpdateStatus(
		c.Context(),
		id,
		"deleted",        // Status baru sesuai instruksi FR-005
		sql.NullString{}, // verified_by kosong
		sql.NullString{}, // rejection_note kosong
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to soft delete achievement",
		})
	}

	// 2. Jika perlu, Anda bisa menambahkan logika tambahan untuk 
	// menyembunyikan data di MongoDB atau menandainya sebagai deleted.

	return c.JSON(fiber.Map{
		"message": "Achievement successfully soft deleted (status changed to deleted)",
	})
}

// Submit godoc
// @Summary      Submit achievement
// @Description  Mengajukan prestasi untuk diverifikasi oleh dosen (FR-004)
// @Tags         Achievements
// @Param        id   path      string  true  "Achievement UUID"
// @Produce      json
// @Success      200  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/achievements/{id}/submit [post]
func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	return s.PgRepo.UpdateStatus(
		c.Context(),
		id,
		"submitted",
		sql.NullString{},
		sql.NullString{},
	)
}

// Verify godoc
// @Summary      Verify achievement
// @Description  Dosen Wali menyetujui prestasi mahasiswa (FR-007)
// @Tags         Achievements
// @Param        id   path      string  true  "Achievement UUID"
// @Produce      json
// @Success      200  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/achievements/{id}/verify [post]
func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	return s.PgRepo.UpdateStatus(
		c.Context(),
		id,
		"verified",
		sql.NullString{String: "lecturer", Valid: true},
		sql.NullString{},
	)
}

// Reject godoc
// @Summary      Reject achievement
// @Description  Dosen Wali menolak prestasi dengan catatan (FR-008)
// @Tags         Achievements
// @Param        id    path      string               true  "Achievement UUID"
// @Param        body  body      object{note=string}  true  "Rejection Note"
// @Accept       json
// @Produce      json
// @Success      200   {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/achievements/{id}/reject [post]
func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	var body struct {
		Note string `json:"note"`
	}
	_ = c.BodyParser(&body)

	return s.PgRepo.UpdateStatus(
		c.Context(),
		id,
		"rejected",
		sql.NullString{},
		sql.NullString{String: body.Note, Valid: true},
	)
}

// UploadAttachment godoc
// @Summary      Upload achievement attachment
// @Description  Mengunggah lampiran dokumen bukti prestasi
// @Tags         Achievements
// @Param        id    path      string  true  "Achievement UUID"
// @Param        file  formData  file    true  "Bukti Dokumen"
// @Accept       multipart/form-data
// @Produce      json
// @Success      200   {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /api/v1/achievements/{id}/attachments [post]
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid UUID format"})
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "File is required"})
	}

	// 1. Ambil data dari PG untuk mendapatkan Mongo ID
	ref, err := s.PgRepo.GetByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Achievement not found in PostgreSQL"})
	}

	// 2. Simpan file secara lokal
	uploadDir := "./uploads/achievements"
	_ = os.MkdirAll(uploadDir, 0755)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	path := filepath.Join(uploadDir, filename)
	if err := c.SaveFile(file, path); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to save file"})
	}

	// 3. Update ke MongoDB
	mongoID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	attachment := model.Attachment{
		Filename: filename,
		Url:      "/uploads/achievements/" + filename,
	}

	err = s.MongoRepo.AddAttachment(c.Context(), mongoID, attachment)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update MongoDB"})
	}

	return c.JSON(fiber.Map{"message": "File uploaded successfully", "data": attachment})
}

// GetHistory godoc
// @Summary      Get achievement history
// @Description  Melihat riwayat perubahan status prestasi
// @Tags         Achievements
// @Param        id   path      string  true  "Achievement UUID"
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /api/v1/achievements/{id}/history [get]
func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))
	history, _ := s.PgRepo.GetStatusHistory(c.Context(), id)
	return c.JSON(fiber.Map{"data": history})
}

// GetStatistics godoc
// @Summary      Get achievement statistics
// @Description  Mendapatkan statistik prestasi (FR-011)
// @Tags         Reports
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /api/v1/reports/statistics [get]
func (s *AchievementService) GetStatistics(c *fiber.Ctx) error {
	stats, _ := s.PgRepo.GetStatistics(c.Context())
	return c.JSON(fiber.Map{"data": stats})
}

// GetStudentReport godoc
// @Summary      Get student achievement report
// @Description  Mendapatkan laporan lengkap prestasi per mahasiswa (FR-012)
// @Tags         Reports
// @Param        id   path      string  true  "Student UUID"
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /api/v1/reports/student/{id} [get]
func (s *AchievementService) GetStudentReport(c *fiber.Ctx) error {
	id := c.Params("id")
	report, _ := s.PgRepo.GetStudentReport(c.Context(), id)
	return c.JSON(fiber.Map{"student_id": id, "summary": report})
}