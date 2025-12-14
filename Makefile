.PHONY: generate
generate:
	mkdir -p gen
	protoc \
		--go_out=gen \
		--go_opt=paths=source_relative \
		--go-grpc_out=gen \
		--go-grpc_opt=paths=source_relative \
		--proto_path=api/proto \
		api/proto/ratelimiter/v1/ratelimiter.proto