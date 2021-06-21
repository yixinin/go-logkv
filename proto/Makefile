service.pb.go: service.proto
	protoc *.proto --go_out=plugins=grpc:. --go_opt=paths=source_relative

force:
	rm -f *.pb.go
