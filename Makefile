all:
	go run *.go
	chrome localhost:8080

deploy:
	gcloud app deploy --version=0
