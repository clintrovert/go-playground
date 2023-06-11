# go-backend-playground
Simple go server with Docker build file

Project structure follows [goland-standards](https://github.com/golang-standards/project-layout) recommendations.

```bash
protoc --go_out=. --go_opt=paths=source_relative \  
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/model/*.proto
```

```bash
grpcurl -d '{"user_id":"test"}' -plaintext localhost:9099 playground.UserService.GetUser
```

### For Mongo

`docker-compose -f build/docker-compose-mongo.yml up -d`

`mongosh "mongodb://root:example@localhost:27017"`