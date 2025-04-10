How to use

0. Cmd + Shift + . (to show file .env.dev)
1. Add these files to root of your microservice, for ex. /parsing
2. cd parsing in repository
3. docker-compose --env-file ./.env.dev up (for db)
4. go run cmd/main.go (for backend, we don't do this in docker containers because it takes a lot of time to build image..)