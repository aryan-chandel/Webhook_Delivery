#Webhook Delivery Service made with golang,mongodb,redis,docker and aws..
    Lets start with deployment on local machine ..can be done through 1)cloning this git repo locally and run docker-compose up this will start the service 2)only through docker 
    Assuming docker and docker compose is installed on machine if not kindly install for smooth deployment (assuming ubuntu os):
    follow these steps
    1)run mkdir project  //make a folder to start 
    2)run cd project/    //get inside project
    3)run touch docker-compose.yml  //make a docker compose file
    4)run nano docker-compose.yml //open this file 
    5)paste the following code there and save 
version: "3.8"

services:
  mongo:
    image: mongo:6.0
    container_name: mongo
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: dev
      MONGO_INITDB_ROOT_PASSWORD: test
    volumes:
      - mongo-data:/data/db

  redis:
    image: redis:7.2
    container_name: redis
    ports:
      - "6379:6379"
    command: ["redis-server", "--requirepass", "admin"]
    volumes:
      - redis-data:/data

  app:
    image: aryanchandel93/aryan_chandel_webhook:latest
    container_name: webhook_service
    ports:
      - "8000:8000"
    environment:
      MONGO_URI: "mongodb://dev:test@mongo:27017"
      REDIS_ADDR: "redis:6379"
      REDIS_PASSWORD: "admin"
    depends_on:
      - mongo
      - redis

volumes:
  mongo-data:
  redis-data: //till here

  6)now exit and run docker-compose up --build //this will start the service and ready to test 

  **after running here are some sample curl command to test 
  1) curl http://localhost:8000/ping    -->it should give status working means server is up and running
  2) curl  curl -X POST http://localhost:8000/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
        "target_url": "https://example.com/webhook-receiver"
      }'    --> it will give back success message and subscription id save that for future use

   3) curl http://localhost:8000/subscriptions/<subscription_id> -->on using same subscription id you got you will get generated subscription from mongodb
   4) now for ingestion testing first run -->export PAYLOAD='{"orderId":1234,"status":"created"}'
then-->
curl -i -X POST http://localhost:8000/ingest/<SUB_ID_you_saved> \
  -H "Content-Type: application/json" \
  -H "Event-Type: order.created" \
  -d "$PAYLOAD"  --> this will fire a go routine that will try to do delivery based on target url

5) curl http://localhost:8000/subscription/logs/<subscription_id> -->this will give the logs of intended subscriber
6) curl -X DELETE http://localhost:8000/subscriptions/<subscription_id> --> this will show success message of deletion of subscriber

   **LIVE LINK OF AWS DEPLOYED PROJECT:
   public dns :ec2-54-163-172-191.compute-1.amazonaws.com  //just use this dns in place of localhost like curl http://ec2-54-163-172-191.compute-1.amazonaws.com:8000/ping and all set you can test like that
   public ipv4: 54.163.172.191 //you can use this ip in place of dns also..
   You can also test all endpoints through postman using these public dns or ip..

***ABOUT ARCHITECTURE AND DATABASE :
    - **Framework**: This application is built using the [Echo](https://echo.labstack.com/) framework for Go, which is lightweight, fast, and provides middleware support.
                     Also echo provides more easier functions and libraries to handle http requests and responses effectively.
                     Another reason for choosing is easy threading just use go keyword and boom no more bottlenecks, waiting time ..New thread is generated.
    - **Database**: MongoDB is used as the primary database to store subscription details and delivery logs.
                    Due to its nosql architecture and our schema less user data it is best choice.
                    Secondary REDIS is used for caching data and handling background worker to complete delivery tasks..
    - **Async Task/Queueing System**: Redis is used to handle queuing for the webhook deliveries. This ensures that the delivery attempts can be retried asynchronously in case of failure.
    - **Retry Strategy**: The system retries webhook delivery up to a specified maximum number of attempts, with a delay between each attempt. This is managed through Redis as the queue.
    - **Log Retention : Also setup a background go routine to keep deleting logs older than 72 hours.. 
### Database Schema & Indexing Strategy

#### Collections
- **Subscriptions**: Contains details about each subscription (e.g., event type, secret, address).
- **Logs**: Stores logs for each delivery attempt, including the webhook ID, response status, and any failure messages.

#### Indexing:
- Indexing is applied to frequently queried fields, such as:
  - `subscription_id`
  - `webhook_id`
  - `status`
  
MongoDB indexes ensure efficient searching for logs and status retrieval.

###Estimated Monthly Cost
    AWS Free Tier
    Assuming the following services:
    EC2 Instance (t2.micro): Free tier allows 750 hours/month.
    S3 (for storage, if used): Free tier includes 5GB.
    MongoDB: MongoDB Atlas offers a free tier with 512MB storage.
            Or mongodb is running locally on ec2 machine under free tier time range 
    Redis: Redis can be run on EC2, which is part of the free tier for t2.micro.

Assumptions:

5000 Webhooks/day.
1.2 delivery attempts/webhook.
The monthly cost would be approximately $0, assuming the usage fits within the free tier limits for EC2, S3.
However, if the usage goes beyond the free tier (e.g., higher traffic or storage), the cost can grow depending on the AWS services' pricing.


### 'Final note'
-A minimal ui also deployed on aws ec2 through Apache server which can be accessed through provided public dns or ip.
-it is made with html ,tailwind css and javascript and with some help of ai tool like chat-GPT..
-I have not included these files in main project repo to keep it backend oriented and keep less size of docker image.
-I can share these files feel free to ask..
-However due to lack of javascript knowledge it may behave incorrectly depending on complex requests .

## Implemented all bonus points as well as required points with efficient approach.

-avoid clicking direct links from readme file and copy full curl commands.
-I will be runnning my ec2 instance for week or 10 days if anyhow it public dns doesnt work kindly ask me to restart my instance i will do so .. :) 
# Full backend system is built and tested by me if any issues come I can assure it will surely be deployment or any typo error we can resolve feel free to ask :)

Thank you..

    

  
