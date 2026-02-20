package database

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    "github.com/joho/godotenv"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func InitDatabase() {
    // Load .env files (try multiple locations)
    _ = godotenv.Load(".env")
    _ = godotenv.Load("../.env")

    connectMongoDB()
}

func connectMongoDB() {
    uri := os.Getenv("MONGO_URI")
    dbName := os.Getenv("DB_NAME")

    if uri == "" {
        uri = "mongodb://localhost:27017"
    }
    
    if dbName == "" {
        dbName = "jaromind"
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Use the new Connect method instead of deprecated NewClient
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatal("Failed to create mongo client:", err)
    }

    if err := client.Ping(ctx, nil); err != nil {
        log.Fatal("❌ MongoDB ping failed:", err)
    }

    Client = client
    DB = client.Database(dbName)

    fmt.Println("✅ Connected to MongoDB successfully!")
}

func GetCollection(collectionName string) *mongo.Collection {
    if DB == nil {
        InitDatabase()
    }
    return DB.Collection(collectionName)
}

// Add this function for services that need the DB instance
func GetDB() *mongo.Database {
    if DB == nil {
        InitDatabase()
    }
    return DB
}