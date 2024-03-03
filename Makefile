CI-all: ci

PR-approval:
	@echo "Running PR CI"
	go build ./...
	go vet ./...
	go test ./...
ci: clean
	# For each subdirectory of the cmd directory, run make ci
	for d in cmd/*; do \
		(cd $$d && make ci); \
	done
	# Clean up
	make clean
clean:
	@echo "Cleaning up"
	# Loop through all subdirectories of the cmd directory and run make clean
	for d in cmd/*; do \
		(cd $$d && make clean); \
	done
codegen:
	@echo "Generating code"
	go install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest

	cd ./pkg/codegen/apis

	# Loop through each directory. Inside the directory we want to use the directory name as the package name.
	for d in ./pkg/codegen/apis/*; do \
		(cd $$d && go generate); \
	done
