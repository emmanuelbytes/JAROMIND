// scripts/setup_tutor_indexes.go
//
// Run once to create indexes for the tutors and tutor_bookings collections.
//
// Usage:
//   go run scripts/setup_tutor_indexes.go
//
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		log.Fatal("MONGODB_URI not set")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "jaromind"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(dbName)

	// ── tutors collection ─────────────────────────────────────────────────────

	tutors := db.Collection("tutors")

	tutorIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "tutorId", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("tutorId_unique"),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("email_unique"),
		},
		{
			Keys:    bson.D{{Key: "isActive", Value: 1}, {Key: "rating", Value: -1}},
			Options: options.Index().SetName("active_rating"),
		},
		{
			Keys:    bson.D{{Key: "subjects", Value: 1}},
			Options: options.Index().SetName("subjects"),
		},
		{
			// Text index for search across name, bio, subjects, tags
			Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "bio", Value: "text"},
				{Key: "subjects", Value: "text"},
				{Key: "tags", Value: "text"},
			},
			Options: options.Index().SetName("tutor_text_search"),
		},
	}

	tutorRes, err := tutors.Indexes().CreateMany(ctx, tutorIndexes)
	if err != nil {
		log.Printf("Warning: some tutor indexes may already exist: %v", err)
	} else {
		fmt.Printf("✅ Tutor indexes created: %v\n", tutorRes)
	}

	// ── tutor_bookings collection ─────────────────────────────────────────────

	bookings := db.Collection("tutor_bookings")

	bookingIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "bookingId", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("bookingId_unique"),
		},
		{
			// Enforce one booking per user-tutor-date-slot combination
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "tutorId", Value: 1},
				{Key: "sessionDate", Value: 1},
				{Key: "timeSlot", Value: 1},
			},
			Options: options.Index().
				SetSparse(true).
				SetName("user_tutor_slot_unique"),
			// Note: not SetUnique here because cancelled bookings
			// should allow re-booking the same slot; enforce uniqueness
			// at the application layer for status "pending"/"confirmed".
		},
		{
			// Tutor availability check query
			Keys: bson.D{
				{Key: "tutorId", Value: 1},
				{Key: "sessionDate", Value: 1},
				{Key: "status", Value: 1},
			},
			Options: options.Index().SetName("tutor_date_status"),
		},
		{
			// User's booking list
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "sessionDate", Value: -1},
			},
			Options: options.Index().SetName("user_bookings"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("status"),
		},
		{
			Keys:    bson.D{{Key: "paymentStatus", Value: 1}},
			Options: options.Index().SetName("paymentStatus"),
		},
	}

	bookingRes, err := bookings.Indexes().CreateMany(ctx, bookingIndexes)
	if err != nil {
		log.Printf("Warning: some booking indexes may already exist: %v", err)
	} else {
		fmt.Printf("✅ Booking indexes created: %v\n", bookingRes)
	}

	fmt.Println("\n✅ Tutor index setup complete.")
}