# go-backend-playground
Simple go server with Docker build file

Project structure follows [goland-standards](https://github.com/golang-standards/project-layout) recommendations.

### Pre-requisites

Generate modeling structs for gRPC contract

```bash
protoc --go_out=. --go_opt=paths=source_relative \  
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/model/*.proto
```

### Local Development

Requesting unary User endpoints - 

#### Get User
```bash
grpcurl -d '{"id":"<test>"}' -plaintext localhost:9099 playground.UserService.GetUser
```

#### Create User
```bash
grpcurl -d '{"name":"test","email":"test@test.com","password":"helloworld"}' -plaintext localhost:9099 playground.UserService.CreateUser
```

#### Update User
```bash
grpcurl -d '{"id":"<test>","name":"updatedName","email":"updatedEmail@password.com", "password":"updatedPassword"}' -plaintext localhost:9099 playground.UserService.UpdateUser
```

#### Delete User
```bash
grpcurl -d '{"user_id":"<test>"}' -plaintext localhost:9099 playground.UserService.DeleteUser
```

### For Mongo

`docker-compose up -d`

`mongosh "mongodb://root:example@localhost:27017"`