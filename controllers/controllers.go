package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
	"webhook_delivery/database"
	"webhook_delivery/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var SubscriberCollection = database.SubscriberData(database.Client, "Subscriber")
var DeliveryCollection = database.DeliveryData(database.Client, "Delivery")

func CacheSubscriptionDetails(subscriptionID string, subscription *models.Subscriber) error {
	var ctx = context.Background()
	data, err := json.Marshal(subscription)
	if err != nil {
		log.Println("error marshalling subscription data:", err)
		return err
	}
	err = database.RedisClient.Set(ctx, "sub_cache:"+subscriptionID, data, 15*time.Minute).Err()
	if err != nil {
		log.Println("error setting subscription data in cache:", err)
		return err
	}
	return nil
}

func GetSubscriptionByID(sub_id string) (*models.Subscriber, error) {
	//implement caching here
	//get from redis
	var sub models.Subscriber
	var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	val, err := database.RedisClient.Get(ctx, "sub_cache:"+sub_id).Result()
	if err == nil {
		err = json.Unmarshal([]byte(val), &sub)
		if err != nil {
			log.Println("subscription cannot be unmarshalled", err)

		}
		log.Println("subcription fetched from cache")
		return &sub, nil
	}
	//if not found in redis then query mongo

	filter := bson.M{"sub_id": sub_id}

	err = SubscriberCollection.FindOne(ctx, filter).Decode(&sub)
	if err != nil {
		return nil, err
	}
	//cache the subscription and return it
	err = CacheSubscriptionDetails(sub_id, &sub)
	if err != nil {
		log.Println("cannot cache subscription", err)
	}
	return &sub, nil
}

func AddSubscriber() echo.HandlerFunc {
	return func(c echo.Context) error {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var sub models.Subscriber //generate new subscriber doc
		err := c.Bind(&sub)
		if err != nil {
			log.Println("subscriber binding fails")
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "cannot bind user check your fields"})
		}
		if sub.TargetURL==""{
			log.Println("target url missing")
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "target_url is required"})
		}
		sub.ID = primitive.NewObjectID()
		sub.CreatedAt = time.Now()
		sub.UpdatedAt = time.Now()
		sub.SubscriptionID = uuid.NewString()
		//insert into collection and check error
		_, er := SubscriberCollection.InsertOne(ctx, sub)
		if er != nil {
			log.Println("new subscriber insertion failed")
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "cannot create subsription try again later"})

		}
		log.Println("successfully created subscription")
		return c.JSON(http.StatusOK, echo.Map{"message": "subscription created successfully", "subscription_id": sub.SubscriptionID})
	}

}

func GetSubscriber() echo.HandlerFunc {
	return func(c echo.Context) error {
		sub_id := c.Param("id")
		if sub_id == "" {
			log.Println("id is missing in url")
			return c.JSON(http.StatusNotFound, echo.Map{"error": "subscription id not found"})
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.M{"sub_id": sub_id}
		var sub models.Subscriber

		err := SubscriberCollection.FindOne(ctx, filter).Decode(&sub)
		if err != nil {
			log.Println("cannot find subscription")
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "cannot find subscriber"})
		}
		log.Println("found subscription")
		return c.JSON(http.StatusOK, sub)

	}

}

func UpdateSubscriber() echo.HandlerFunc {
	return func(c echo.Context) error {
		sub_id := c.Param("id")
		if sub_id == "" {
			log.Println("id is missing in url")
			return c.JSON(http.StatusNotFound, echo.Map{"error": "id not found"})
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var updatedata map[string]interface{}
		er := c.Bind(&updatedata)
		if er != nil {
			log.Println("cannot decode required field to be updated")
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid body"})
		}

		filter := bson.M{"sub_id": sub_id}
		update := bson.M{"$set": updatedata}

		_, err := SubscriberCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Println("subscription updation failed")
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "subscription cant be updated"})
		}
		log.Println("subscription updated")
		return c.JSON(http.StatusOK, echo.Map{"message": "subscription updated"})
	}

}

func DeleteSubscriber() echo.HandlerFunc {
	return func(c echo.Context) error {
		sub_id := c.Param("id")
		if sub_id == "" {
			log.Println("id is missing in url")
			return c.JSON(http.StatusNotFound, echo.Map{"error": "subscription id not found"})
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.M{"sub_id": sub_id}
		// find and delete
		_, err := SubscriberCollection.DeleteOne(ctx, filter)
		if err != nil {
			log.Println("cannot delete subscription")
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "cannnot delete subscription"})
		}
		log.Println("subscription deleted")
		return c.JSON(http.StatusOK, echo.Map{"message": "subscription deleted successfully"})
	}

}

func SubscriberStatus() echo.HandlerFunc {
	return func(c echo.Context) error {
		webhook_id := c.Param("id")
		if webhook_id == "" {
			log.Println("id is missing in url")
			return c.JSON(http.StatusNotFound, echo.Map{"error": "id not found"})
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		var newlog models.DeliveryLog
		filter := bson.M{"webhook_id": webhook_id}

		err := DeliveryCollection.FindOne(ctx, filter).Decode(&newlog)
		if err != nil {
			log.Println("cannot find required webhook", err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "something went wrong"})
		}
		log.Println("successfully found webhook log")
		return c.JSON(http.StatusOK, newlog)

	}

}

func SubscriberLog() echo.HandlerFunc {
	return func(c echo.Context) error {
		sub_id := c.Param("id")
		if sub_id == "" {
			log.Println("id is missing in url")
			return c.JSON(http.StatusNotFound, echo.Map{"error": "id is missing"})
		}
		filter := bson.M{"subscription_id": sub_id}
		opts := options.Find().SetLimit(20).SetSort(bson.M{"attempted_at": -1})
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		cursor, err := DeliveryCollection.Find(ctx, filter, opts)
		if err != nil {
			log.Println("cannot find delivery logs")
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "something went wrong"})
		}
		var Logs []models.DeliveryLog
		err = cursor.All(ctx, &Logs)
		if err != nil {
			log.Println("cursor error")
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "something went wrong"})
		}
		return c.JSON(http.StatusOK, Logs)

	}

}

// handle webhook request
func NewDelivery() echo.HandlerFunc {
	return func(c echo.Context) error {
		sub_id := c.Param("id")
		if sub_id == "" {
			log.Println("subscription id is missing in URL")
			return c.JSON(http.StatusNotFound, echo.Map{"error": "id is missing "})
		}
		eventType := c.Request().Header.Get("Event-Type") //  Get event type
        if eventType == "" {
			log.Println("event type missing ")
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Event-Type header missing"})
        }
		payload, er := io.ReadAll(c.Request().Body)
		if er != nil {
			log.Println("task body missing")
			return c.JSON(http.StatusNotFound, echo.Map{"error": "task body is missing"})
		}
		c.Request().Body = io.NopCloser(bytes.NewBuffer(payload)) // reset for future reading

		sub, err := GetSubscriptionByID(sub_id)
		if err != nil {
			log.Println("subscriber not found")
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "subscription not found"})
		}
		allowed := false
        if len(sub.EventTypes) == 0 {
            allowed = true // If no event types specified, allow all
        } else {
            for _, event := range sub.EventTypes {
                if event == eventType {
                    allowed = true
                    break
                }
            }
        }

        if !allowed {
            return c.JSON(http.StatusForbidden, map[string]string{"error": "event type not allowed for this subscription"})
        }
		//create new delivery
		var NewDel models.DeliveryLog
		NewDel.ID = primitive.NewObjectID()
		NewDel.SubscriptionID = sub_id
		NewDel.TargetURL = sub.TargetURL
		NewDel.WebhookID = uuid.New().String()
		NewDel.AttemptNumber = 1
		NewDel.Outcome = "pending"

		//build redis task
		var task models.DeliveryTask
		task.LogID = NewDel.ID
		task.SubscriptionID = NewDel.SubscriptionID
		task.WebhookID = NewDel.WebhookID
		task.AttemptNumber = NewDel.AttemptNumber
		task.TargetURL = sub.TargetURL
		task.AttemptNumber = NewDel.AttemptNumber
		task.Payload = payload

		taskJson, err := json.Marshal(task)
		if err != nil {
			log.Println("cannot marshal task")
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "something went wrong try again later"})
		}
		//now new delivery is ready to insert in Redis queue and collection
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		_, err = DeliveryCollection.InsertOne(ctx, NewDel)
		if err != nil {
			log.Println("Delivery insertion failed")
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Delivery insertion failed"})
		}
		log.Println("delivery insertion successful")

		// enqueue in redis
		err = database.RedisClient.RPush(ctx, "delivery_queue", taskJson).Err()
		if err != nil {
			log.Println(err)
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "cannot enqueue task "})
		}
		log.Println("successfuly created and enqueued webhook")
		return c.JSON(http.StatusOK, echo.Map{"message": "successfully webhook queued", "webhook_id": NewDel.WebhookID})

	}

}
