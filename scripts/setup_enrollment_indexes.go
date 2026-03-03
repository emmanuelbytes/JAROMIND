// scripts/setup_enrollment_indexes.go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/*
	This script creates necessary indexes for the enrollment system
	Run once after deploying the new enrollment system
	
	Usage: go run scripts/setup_enrollment_indexes.go
*/

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update with your MongoDB URI
	mongoURI := "mongodb+srv://emmanuelabara265_db_user:yPpWDSl2v8dyHBsO@cluster0.dvkpuqs.mongodb.net/?appName=Cluster0" // Change this to your MongoDB URI
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Get database (update with your database name)
	db := client.Database("emmanuelabara265_db_user") // Change to your database name
	enrollmentsCollection := db.Collection("enrollments")

	fmt.Println("🔧 Setting up enrollment indexes...")

	// Create indexes for better query performance
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "userId", Value: 1}, {Key: "courseId", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("unique_user_course"),
		},
		{
			Keys: bson.D{{Key: "enrollmentId", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("unique_enrollment_id"),
		},
		{
			Keys: bson.D{{Key: "userId", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("user_status"),
		},
		{
			Keys: bson.D{{Key: "courseId", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("course_status"),
		},
		{
			Keys: bson.D{{Key: "paymentStatus", Value: 1}},
			Options: options.Index().SetName("payment_status"),
		},
		{
			Keys: bson.D{{Key: "enrolledAt", Value: -1}},
			Options: options.Index().SetName("enrolled_at_desc"),
		},
		{
			Keys: bson.D{{Key: "lastAccessedAt", Value: -1}},
			Options: options.Index().SetName("last_accessed_desc"),
		},
	}

	// Create indexes
	indexNames, err := enrollmentsCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("⚠️  Warning: Some indexes might already exist: %v", err)
	} else {
		fmt.Println("✅ Created indexes:")
		for _, name := range indexNames {
			fmt.Printf("   - %s\n", name)
		}
	}

	// Verify indexes
	cursor, err := enrollmentsCollection.Indexes().List(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("\n📋 Current indexes on enrollments collection:")
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Fatal(err)
	}

	for _, result := range results {
		fmt.Printf("   - %v\n", result["name"])
	}

	fmt.Println("\n✅ Enrollment system indexes setup complete!")
	fmt.Println("💡 Note: The unique_user_course index prevents duplicate enrollments")
}