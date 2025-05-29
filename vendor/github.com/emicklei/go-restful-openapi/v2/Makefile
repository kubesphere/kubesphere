clean:
	rm coverage.out

test:
	go test -cover ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# for local testing
lint:
	docker run --rm -v $(PWD):$(PWD) -w $(PWD) golangci/golangci-lint:v1.33.0 golangci-lint run -v

# for testing on CI/CD. we specify required linter version in the .travis.yml file
lint-ci:
	golangci-lint run

outdated:
	go list -u -m -json all | docker run -i psampaz/go-mod-outdated -update -direct -ci