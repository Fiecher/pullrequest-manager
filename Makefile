OAPI_CODEGEN_OUTPUT := internal/api/codegen_api.go

generate-api:
	go tool oapi-codegen -generate types,server -package api api/openapi.yml > $(OAPI_CODEGEN_OUTPUT)