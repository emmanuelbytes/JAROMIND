package controllers

import (
    "context"
    "math"
    "net/http"
    "strconv"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"

    "github.com/AbaraEmmanuel/jaromind-backend/database"
    "github.com/AbaraEmmanuel/jaromind-backend/models"
)

func getApplicationsCollection() *mongo.Collection {
    return database.DB.Collection("tutor_applications")
}

// ── PUBLIC: Submit Application ──────────────────────────────
// POST /apply/tutor
func SubmitTutorApplication(c *gin.Context) {
    var app models.TutorApplication
    if err := c.ShouldBindJSON(&app); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
        return
    }

    // Check for duplicate email
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    count, _ := getApplicationsCollection().CountDocuments(ctx, bson.M{
        "email":  app.Email,
        "status": bson.M{"$in": []string{"pending", "approved"}},
    })
    if count > 0 {
        c.JSON(http.StatusConflict, gin.H{
            "success": false,
            "error":   "An application with this email is already pending or approved.",
        })
        return
    }

    app.ID        = primitive.NewObjectID()
    app.AppID     = uuid.New().String()
    app.Status    = "pending"
    app.CreatedAt = time.Now()
    app.UpdatedAt = time.Now()

    _, err := getApplicationsCollection().InsertOne(ctx, app)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to submit application."})
        return
    }

    // TODO: send confirmation email to app.Email

    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "message": "Application submitted successfully. You'll hear back within 3–5 business days.",
        "appId":   app.AppID,
    })
}

// ── ADMIN: List Applications ─────────────────────────────────
// GET /admin/applications?status=pending&page=1&limit=20
func AdminGetApplications(c *gin.Context) {
    status := c.DefaultQuery("status", "pending")
    page,  _ := strconv.Atoi(c.DefaultQuery("page",  "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    if page < 1  { page  = 1  }
    if limit < 1 { limit = 20 }

    filter := bson.M{}
    if status != "all" {
        filter["status"] = status
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    total, _ := getApplicationsCollection().CountDocuments(ctx, filter)

    opts := options.Find().
        SetSort(bson.D{{Key: "createdAt", Value: -1}}).
        SetSkip(int64((page - 1) * limit)).
        SetLimit(int64(limit))

    cursor, err := getApplicationsCollection().Find(ctx, filter, opts)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to fetch applications."})
        return
    }
    defer cursor.Close(ctx)

    var apps []models.TutorApplication
    cursor.All(ctx, &apps)
    if apps == nil { apps = []models.TutorApplication{} }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    apps,
        "total":   total,
        "page":    page,
        "pages":   int(math.Ceil(float64(total) / float64(limit))),
    })
}

// ── ADMIN: Get Single Application ────────────────────────────
// GET /admin/applications/:id
func AdminGetApplication(c *gin.Context) {
    id  := c.Param("id")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    filter := bson.M{"$or": []bson.M{{"appId": id}}}
    if oid, err := primitive.ObjectIDFromHex(id); err == nil {
        filter = bson.M{"$or": []bson.M{{"appId": id}, {"_id": oid}}}
    }

    var app models.TutorApplication
    if err := getApplicationsCollection().FindOne(ctx, filter).Decode(&app); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Application not found."})
        return
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "data": app})
}

// ── ADMIN: Review (Approve / Reject / Request Revision) ──────
// PUT /admin/applications/:id/review
func AdminReviewApplication(c *gin.Context) {
    id := c.Param("id")
    var req models.ApplicationReviewRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    filter := bson.M{"$or": []bson.M{{"appId": id}}}
    if oid, err := primitive.ObjectIDFromHex(id); err == nil {
        filter = bson.M{"$or": []bson.M{{"appId": id}, {"_id": oid}}}
    }

    // Fetch application first
    var app models.TutorApplication
    if err := getApplicationsCollection().FindOne(ctx, filter).Decode(&app); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Application not found."})
        return
    }

    now         := time.Now()
    adminEmail  := c.GetString("userEmail") // set by JWT middleware

    update := bson.M{"$set": bson.M{
        "status":       req.Status,
        "reviewerNote": req.ReviewerNote,
        "reviewedAt":   now,
        "reviewedBy":   adminEmail,
        "updatedAt":    now,
    }}

    _, err := getApplicationsCollection().UpdateOne(ctx, filter, update)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update application."})
        return
    }

    // ── On APPROVAL: create TutorProfile ──
    if req.Status == "approved" {
        profile := models.TutorProfile{
            ID:           primitive.NewObjectID(),
            TutorID:      uuid.New().String(),
            Name:         app.FirstName + " " + app.LastName,
            Email:        app.Email,
            Phone:        app.Phone,
            AvatarURL:    app.PhotoURL,
            Bio:          app.Bio,
            Subjects:     app.Subjects,
            HourlyRate:   app.HourlyRate,
            IsOnline:     false,
            IsActive:     true,
            Rating:       0,
            ReviewCount:  0,
            SessionCount: 0,
            CreatedAt:    now,
            UpdatedAt:    now,
        }

        getTutorsCollection().InsertOne(ctx, profile)

        // TODO: send approval email to app.Email with login instructions
    }

    // TODO: send rejection/revision email with req.ReviewerNote

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Application " + req.Status + " successfully.",
    })
}

// ── ADMIN: Dashboard counts ───────────────────────────────────
// GET /admin/applications/stats
func AdminApplicationStats(c *gin.Context) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    statuses := []string{"pending", "approved", "rejected", "revision_requested"}
    result   := gin.H{}

    for _, s := range statuses {
        count, _ := getApplicationsCollection().CountDocuments(ctx, bson.M{"status": s})
        result[s] = count
    }

    c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}