package service

import (
	"context"
	"database/sql"
	"log"
	"fmt"
    
	"uas/app/model" 
	"uas/app/repository" 
    
	"go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "github.com/gofiber/fiber/v2"
)

// AchievementService mengimplementasikan logika bisnis dan orkestrasi DB
type AchievementService struct {
    PGAchieveRepo *repository.AchievementRepository
    MongoAchieveRepo *repository.MongoAchievementRepository
}

// =========================================================================
// CRUD Operations
// =========================================================================

// Create (POST /achievements)
func (s *AchievementService) Create(c *fiber.Ctx) error {
    var req model.AchievementCreateUpdate
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }
    if req.StudentID == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Student ID is required"})
    }

    ctx := context.Background()
    
    // 1. Simpan Detail ke MongoDB
    mongoAch := model.MongoAchievement{
        StudentID: req.StudentID,
        AchievementType: req.AchievementType,
        Title: req.Title,
        Description: req.Description,
        Points: req.Points,
        Tags: req.Tags,
        Details: req.Details,
        Attachments: req.Attachments,
    }
    mongoID, err := s.MongoAchieveRepo.Create(ctx, &mongoAch)
    if err != nil {
        log.Printf("Mongo Create Error: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save detail in MongoDB"})
    }

    // 2. Simpan Referensi ke PostgreSQL (Status Awal: draft)
    pgRef := model.AchievementReference{
        StudentID: req.StudentID,
        MongoAchievementID: mongoID.Hex(), 
        Status: "draft", 
    }
    pgID, err := s.PGAchieveRepo.Create(ctx, &pgRef) 
    if err != nil {
        // Rollback: HAPUS DOKUMEN DARI MONGODB JIKA PG GAGAL
        log.Printf("PG Create Error, rolling back Mongo: %v", err)
        s.MongoAchieveRepo.Delete(ctx, mongoID) 
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create reference in PostgreSQL. MongoDB rolled back."})
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "message": "Achievement created successfully and set to draft",
        "achievement_ref_id": pgID,
        "mongo_id": mongoID.Hex(),
        "status": "draft",
    })
}

// GetAll (GET /achievements) - Mengambil semua referensi dari PG
func (s *AchievementService) GetAll(c *fiber.Ctx) error {
    ctx := context.Background()
    
    // Asumsi: Repository PG memiliki GetAll yang mengembalikan []model.AchievementReference
    refs, err := s.PGAchieveRepo.GetAll(ctx)
    if err != nil {
        log.Printf("PG GetAll Error: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch achievement list"})
    }
    
    // Dalam implementasi nyata, Anda harus melakukan JOIN data Mongo di sini
    return c.JSON(refs)
}

// GetDetail (GET /achievements/:id)
func (s *AchievementService) GetDetail(c *fiber.Ctx) error {
    pgID := c.Params("id") 
    ctx := context.Background()

    // 1. Ambil referensi dari PG
    pgRef, err := s.PGAchieveRepo.GetByID(ctx, pgID)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Achievement reference not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch PG reference"})
    }

    // 2. Ambil detail dari Mongo menggunakan Mongo ID
    mongoID, _ := primitive.ObjectIDFromHex(pgRef.MongoAchievementID)
    mongoDetail, err := s.MongoAchieveRepo.GetByID(ctx, mongoID)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            log.Printf("Mongo detail missing for PG ID %s", pgID)
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Achievement detail missing in MongoDB"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch Mongo detail"})
    }
    
    // 3. Gabungkan dan kembalikan
    fullData := model.AchievementFull{
        AchievementReference: pgRef,
        MongoDetails: mongoDetail,
    }

    return c.JSON(fullData)
}

// Update (PUT /achievements/:id) - Update di Mongo, update timestamp di PG
func (s *AchievementService) Update(c *fiber.Ctx) error {
    pgID := c.Params("id")
    var req model.AchievementCreateUpdate
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
    }

    ctx := context.Background()
    
    // 1. Ambil ID Mongo dari PG
    pgRef, err := s.PGAchieveRepo.GetByID(ctx, pgID)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Achievement reference not found for update"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch PG reference for update"})
    }

    // 2. Update Detail di MongoDB
    mongoID, _ := primitive.ObjectIDFromHex(pgRef.MongoAchievementID)
    updateData := map[string]interface{}{
        "achievement_type": req.AchievementType,
        "title": req.Title,
        "description": req.Description,
        "points": req.Points,
        "tags": req.Tags,
        "details": req.Details,
        "attachments": req.Attachments,
    }

    if err := s.MongoAchieveRepo.Update(ctx, mongoID, updateData); err != nil {
        log.Printf("Mongo Update Error: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update detail in MongoDB"})
    }
    
    // 3. Update timestamp di PG (Asumsi: PGAchieveRepo punya UpdateTimestamp)
    if err := s.PGAchieveRepo.UpdateTimestamp(ctx, pgID); err != nil {
        log.Printf("PG Timestamp Update Error: %v", err)
        // Non-fatal error, tapi perlu log
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Achievement updated successfully"})
}

// Delete (DELETE /achievements/:id) - Hapus dari PG lalu Mongo
func (s *AchievementService) Delete(c *fiber.Ctx) error {
    pgID := c.Params("id")
    ctx := context.Background()

    // 1. Ambil ID Mongo dari PG
    pgRef, err := s.PGAchieveRepo.GetByID(ctx, pgID)
    if err != nil {
        if err == sql.ErrNoRows {
            return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Achievement reference not found for deletion"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch PG reference for deletion"})
    }

    // 2. Hapus referensi dari PostgreSQL
    if err := s.PGAchieveRepo.Delete(ctx, pgID); err != nil {
        log.Printf("PG Delete Error: %v", err)
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete reference from PostgreSQL"})
    }

    // 3. Hapus detail dari MongoDB (Asumsi: MongoAchieveRepo.Delete ada)
    mongoID, _ := primitive.ObjectIDFromHex(pgRef.MongoAchievementID)
    if err := s.MongoAchieveRepo.Delete(ctx, mongoID); err != nil {
        log.Printf("Mongo Delete Error (after PG delete): %v", err)
        // Fatal, tapi karena PG sudah terhapus, anggap sukses, log warning
    }

    return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Achievement deleted successfully"})
}

// =========================================================================
// Workflow Operations
// =========================================================================

// Submit (POST /achievements/:id/submit) - Update status di PG
func (s *AchievementService) Submit(c *fiber.Ctx) error {
    pgID := c.Params("id") 
    ctx := context.Background()

    // Update status di PG dari 'draft' ke 'submitted'
    err := s.PGAchieveRepo.UpdateStatus(
        ctx, 
        pgID, 
        "submitted", 
        sql.NullString{}, // verifiedBy = NULL
        sql.NullString{}, // rejectionNote = NULL
    )
    if err != nil {
        if err == sql.ErrNoRows {
             return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Achievement reference not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to submit achievement: %v", err)})
    }
    
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": fmt.Sprintf("Achievement %s successfully submitted for verification.", pgID),
        "new_status": "submitted",
    })
}

// Verify (POST /achievements/:id/verify) - Update status di PG menjadi verified
func (s *AchievementService) Verify(c *fiber.Ctx) error {
    pgID := c.Params("id") 
    // Ambil User ID dari Token JWT (Asumsi: c.Locals("user_id") berisi UUID User)
    verifierID := c.Locals("user_id").(string) 
    ctx := context.Background()

    // Update status di PG menjadi 'verified'
    err := s.PGAchieveRepo.UpdateStatus(
        ctx, 
        pgID, 
        "verified", 
        sql.NullString{String: verifierID, Valid: true}, // verifiedBy = User ID
        sql.NullString{}, 
    )
    if err != nil {
        if err == sql.ErrNoRows {
             return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Achievement reference not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to verify achievement: %v", err)})
    }
    
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": fmt.Sprintf("Achievement %s successfully verified.", pgID),
        "new_status": "verified",
    })
}

// Reject (POST /achievements/:id/reject) - Update status di PG menjadi rejected
func (s *AchievementService) Reject(c *fiber.Ctx) error {
    pgID := c.Params("id") 
    verifierID := c.Locals("user_id").(string)
    
    var req struct {
        RejectionNote string `json:"rejection_note"`
    }
    if err := c.BodyParser(&req); err != nil || req.RejectionNote == "" {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Rejection note is required for rejection"})
    }

    ctx := context.Background()

    // Update status di PG menjadi 'rejected'
    err := s.PGAchieveRepo.UpdateStatus(
        ctx, 
        pgID, 
        "rejected", 
        sql.NullString{String: verifierID, Valid: true}, 
        sql.NullString{String: req.RejectionNote, Valid: true}, 
    )
    if err != nil {
        if err == sql.ErrNoRows {
             return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Achievement reference not found"})
        }
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to reject achievement: %v", err)})
    }
    
    return c.Status(fiber.StatusOK).JSON(fiber.Map{
        "message": fmt.Sprintf("Achievement %s successfully rejected.", pgID),
        "new_status": "rejected",
        "note": req.RejectionNote,
    })
}

// =========================================================================
// File & History Operations
// =========================================================================

// UploadAttachment (POST /achievements/:id/attachments)
func (s *AchievementService) UploadAttachment(c *fiber.Ctx) error {
    // Logika: 
    // 1. Ambil file dari request (c.FormFile)
    // 2. Simpan file ke storage (S3/Local) dan dapatkan URL
    // 3. Panggil MongoAchieveRepo.AddAttachment untuk push ke array attachments di Mongo
    return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "UploadAttachment logic not yet implemented"})
}

// GetHistory (GET /achievements/:id/history)
func (s *AchievementService) GetHistory(c *fiber.Ctx) error {
    // Logika: 
    // 1. Ambil data historis dari tabel audit/log di PG berdasarkan achievement_references.id
    return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "GetHistory logic not yet implemented"})
}

// =========================================================================
// REPORTS & ANALYTICS
// =========================================================================

// GetStatistics (GET /reports/statistics)
func (s *AchievementService) GetStatistics(c *fiber.Ctx) error {
    // Logika: 
    // 1. Panggil PGAchieveRepo.GetStatusCounts untuk menghitung total draft, submitted, verified, rejected.
    return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "GetStatistics logic not yet implemented"})
}

// GetStudentReport (GET /reports/student/:id)
func (s *AchievementService) GetStudentReport(c *fiber.Ctx) error {
    // Logika:
    // 1. Ambil semua referensi prestasi mahasiswa (:id adalah student_id) dari PG
    // 2. Iterasi dan ambil detail Mongo untuk setiap referensi (Mirip GetAll tapi dengan Filter)
    return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "GetStudentReport logic not yet implemented"})
}

func NewAchievementService(
    pgRepo *repository.AchievementRepository, 
    mongoRepo *repository.MongoAchievementRepository,
) *AchievementService {
    return &AchievementService{
        PGAchieveRepo: pgRepo,
        MongoAchieveRepo: mongoRepo,
    }
}