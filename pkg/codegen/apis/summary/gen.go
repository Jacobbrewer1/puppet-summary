package summary

//go:generate oapi-codegen -generate types -package summary -templates ../../templates -o types.go ./routes.yaml
//go:generate oapi-codegen -generate gorilla -package summary -templates ../../templates -o server.go ./routes.yaml
