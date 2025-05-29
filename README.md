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


    

  
