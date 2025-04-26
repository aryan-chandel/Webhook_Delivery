package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"
	"webhook_delivery/controllers"
	"webhook_delivery/database"
	"webhook_delivery/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ctx = context.Background()
var client = http.Client{
	Timeout: 10 * time.Second,
}
var ErrStatusNotOkay = errors.New("response status not okay")

func StartWorker() {
	for {
		taskdata, err := database.RedisClient.BLPop(ctx, 0, "delivery_queue").Result()
		if err != nil {
			log.Println("error during extraction of task from queue", err)
			continue
		}
		var task models.DeliveryTask
		err = json.Unmarshal([]byte(taskdata[1]), &task)
		if err != nil {
			log.Println("error while unmarshalling task", err)
			continue
		}
		err = ProcessTask(&task)
		if err != nil {
			log.Println("error while processing", err)
			//start retry in another go routine so it should not stuck in sleep time
			go RetryTask(&task)
		}

	}
}
func ProcessTask(task *models.DeliveryTask) error {
	//check for max retries
	if task.AttemptNumber > 5 {
		UpdateDeliveryLog(task.LogID, task.AttemptNumber, "failed", http.StatusBadRequest)
		return nil

	}
	//form post request for target url
	req, err := http.NewRequest("POST", task.TargetURL, bytes.NewReader(task.Payload))
	if err != nil {
		log.Println("cannot create request", err)
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("error while sending webhook request", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("non-Ok response", resp.Status)
		err = UpdateDeliveryLog(task.LogID, task.AttemptNumber, "failed this time...Retrying", resp.StatusCode)
		if err != nil {
			log.Println("cannot update log", resp.StatusCode)
			return err

		}
		return ErrStatusNotOkay

	}
	err = UpdateDeliveryLog(task.LogID, task.AttemptNumber, "successful", resp.StatusCode)
	if err != nil {
		log.Println("cannot update log", resp.StatusCode)
		return nil

	}
	log.Println("successfully webhook delivered")
	return nil

}
func RetryTask(task *models.DeliveryTask) {
	//first increase the attempt no. should go max 5
	task.AttemptNumber++
	if task.AttemptNumber > 5 {
		log.Println("Max retry attempts reached. Dropping task.")
		return
	}

	// generate retry period
	retry_time := time.Duration(2*task.AttemptNumber-1) * time.Second
	log.Printf("retry time is set for task as %v", retry_time)

	//hold go routine
	time.Sleep(retry_time)
	//re-enqueue into redis
	taskjson, err := json.Marshal(task)
	if err != nil {
		log.Println("error marshalling the task ", err)
		return
	}
	err = database.RedisClient.RPush(ctx, "delivery_queue", taskjson).Err()
	if err != nil {
		log.Println("cannot re-enqueue task", err)
		return
	}
	log.Println("task re-enqueued successfully")

}

func UpdateDeliveryLog(id primitive.ObjectID, attemptedNo int, outcome string, status int) error {
	var ctxx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"attempted_number": attemptedNo, "outcome": outcome, "attempted_at": time.Now(), "http_status": status}}
	_, err := controllers.DeliveryCollection.UpdateOne(ctxx, filter, update)
	if err != nil {
		log.Println("cannot find log", err)
		return err
	}
	log.Println("successfully updated delivery log")
	return nil
}
