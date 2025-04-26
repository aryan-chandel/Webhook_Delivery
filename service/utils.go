package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"
	"webhook_delivery/controllers"

	"go.mongodb.org/mongo-driver/bson"
)

// ComputeHMAC generates a HMAC-SHA256 signature from the payload and secret.
func ComputeHMAC(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// make a ticker to delete log after every 6 hours
func StartLogRententionWorker() {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		CleanOldLog()
	}
}

func CleanOldLog() {
	//generate time stamp for 3 days
	expires := time.Now().Add(-72 * time.Hour)
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"attempted_at": bson.M{"$lt": expires}}
	result, err := controllers.DeliveryCollection.DeleteMany(ctx, filter)
	if err != nil {
		log.Println("cannot delete logs", err)
		return
	}
	log.Printf("successfully deleted %v logs", result.DeletedCount)
}
