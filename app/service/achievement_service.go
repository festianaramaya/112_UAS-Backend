package service

import (
    "time"
    "fmt"
	"os"
	"path/filepath"

	"database/sql"
	"uas/app/model"
	"uas/app/repository"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
    
	"go.mongodb.org/mongo-driver/bson/primitive"
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

func (s *AchievementService) GetAll(c *fiber.Ctx) error {
	data, err := s.PgRepo.GetAll(c.Context())
	if err != nil {
		return err
	}
	return c.JSON(data)
}

func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	data, err := s.PgRepo.GetByID(c.Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(data)
}

func (s *AchievementService) Create(c *fiber.Ctx) error {
	var mongoData model.MongoAchievement

	// parse body
	if err := c.BodyParser(&mongoData); err != nil {
		return fiber.ErrBadRequest
	}

	// validasi student_id harus UUID
	studentUUID, err := uuid.Parse(mongoData.StudentID)
	if err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"student_id must be valid UUID",
		)
	}

	ctx := c.Context()

	// 1️⃣ simpan ke MongoDB
	mongoID, err := s.MongoRepo.Create(ctx, &mongoData)
	if err != nil {
		return err
	}

	// 2️⃣ simpan reference ke PostgreSQL
	ref := model.AchievementReference{
		StudentID:          studentUUID.String(),
		MongoAchievementID: mongoID.Hex(),
		Status:             "draft",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.PgRepo.Create(&ref); err != nil {
		_ = s.MongoRepo.Delete(ctx, mongoID) // rollback
		return err
	}

	// 3️⃣ response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "achievement created",
		"data": fiber.Map{
			"id":                   ref.ID,
			"student_id":           ref.StudentID,
			"mongo_achievement_id": ref.MongoAchievementID,
			"status":               ref.Status,
			"created_at":           ref.CreatedAt,
		},
	})
}

func (s *AchievementService) Update(c *fiber.Ctx) error {
	// STRING ➜ UUID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.ErrBadRequest
	}

	var payload map[string]interface{}
	if err := c.BodyParser(&payload); err != nil {
		return err
	}

	// Update timestamp di PostgreSQL
	if err := s.PgRepo.UpdateTimestamp(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "achievement updated",
	})
}

func (s *AchievementService) Delete(c *fiber.Ctx) error {
	// STRING ➜ UUID
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.ErrBadRequest
	}

	if err := s.PgRepo.Delete(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "achievement deleted",
	})
}


/* ===================== WORKFLOW ===================== */

func (s *AchievementService) Submit(c *fiber.Ctx) error {
	id, _ := uuid.Parse(c.Params("id"))

	return s.PgRepo.UpdateStatus(
		c.Context(),
		id,
		"submitted",
		sql.NullString{},
		sql.NullString{},
	)
}


func (s *AchievementService) Verify(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	err = s.PgRepo.UpdateStatus(
		c.Context(),
		id,
		"verified",
		sql.NullString{String: "lecturer", Valid: true},
		sql.NullString{},
	)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Achievement verified successfully",
		"status":  "verified",
	})
}

func (s *AchievementService) Reject(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	var body struct {
		Note string `json:"note"`
	}
	_ = c.BodyParser(&body)

	err = s.PgRepo.UpdateStatus(
		c.Context(),
		id,
		"rejected",
		sql.NullString{},
		sql.NullString{String: body.Note, Valid: true},
	)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Achievement rejected",
		"status":  "rejected",
		"note":    body.Note,
	})
}



/* ===================== EXTRA FEATURES ===================== */
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
	// ambil achievement id (PG id)
	idStr := c.Params("id")
	_, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.ErrBadRequest
	}

	// ambil file
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}

	// buat folder
	uploadDir := "./uploads/achievements"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return err
	}

	// nama file unik
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
	filePath := filepath.Join(uploadDir, filename)

	// simpan file
	if err := c.SaveFile(file, filePath); err != nil {
		return err
	}

	// ambil mongo_id dari PostgreSQL
	ref, err := s.PgRepo.GetByID(c.Context(), uuid.MustParse(idStr))
	if err != nil {
		return err
	}

	mongoID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return err
	}

	// buat attachment
	attachment := model.Attachment{
		Filename: filename,
		Url:      "/uploads/achievements/" + filename,
	}

	// ⬅️ SIMPAN KE MONGO
	if err := s.MongoRepo.AddAttachment(c.Context(), mongoID, attachment); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "file uploaded",
		"data":    attachment,
	})
}

func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.ErrBadRequest
	}

	history, err := s.PgRepo.GetStatusHistory(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Achievement status history",
		"data":    history,
	})
}

func (s *AchievementService) GetStatistics(c *fiber.Ctx) error {
	stats, err := s.PgRepo.GetStatistics(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Achievement statistics",
		"data":    stats,
	})
}

func (s *AchievementService) GetStudentReport(c *fiber.Ctx) error {
	studentID := c.Params("id")

	report, err := s.PgRepo.GetStudentReport(c.Context(), studentID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Student achievement report",
		"data": fiber.Map{
			"student_id": studentID,
			"summary":    report,
		},
	})
}
