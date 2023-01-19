BINARY_NAME = "go-simple-bank"
BINARIES = "./bin"
MAIN_DIR = "cmd/go-simple-bank/"
GITHUB = "github.com/max-rodziyevsky/go-simple-bank"
GIT_LOCAL_NAME = "max-rodziyevsky"
GIT_LOCAL_EMAIL = "rodziyevskydev@gmail.com"

init:
	@echo "::> Creating a module root..."
	@go mod init ${GITHUB}
	@mkdir "cmd" && mkdir "cmd/"${BINARY_NAME}
	@touch ${MAIN_DIR}/main.go
	@echo "package main\n\nfunc main(){\n\n}" > ${MAIN_DIR}/main.go
	@echo "::> Finished!"

build:
	@echo "::> Building..."
	@go build -o ${BINARIES}/${BINARY_NAME} ${MAIN_DIR}
	@echo "::> Finished!"

run:
	@go build -o ${BINARIES}/${BINARY_NAME} ${MAIN_DIR}
	@${BINARIES}/${BINARY_NAME}

clean:
	@echo "::> Cleaning..."
	@go clean
	@rm -rf ${BINARIES}
	@go mod tidy
	@echo "::> Finished"

local-git:
	@git config --local user.name ${GIT_LOCAL_NAME}
	@git config --local user.email ${GIT_LOCAL_EMAIL}
	@git config --local --list

git-init:
	@echo "::> Git initialization begin..."
	@git init
	@git config --local user.name ${GIT_LOCAL_NAME}
	@git config --local user.email ${GIT_LOCAL_EMAIL}
	@touch .gitignore
	@echo ".idea" > .gitignore
	@echo "bin" > .gitignore
	@touch README.md
	@git add README.md
	@git commit -m "first commit"
	@git branch -M main
	@git remote add origin https://${GITHUB}
	@git push -u origin main
	@echo "::> Finished"

## Database operations
postgres:
	docker run --name go-simple-bank-db -p 5433:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:15.1-alpine

create-db:
	docker exec -it go-simple-bank-db createdb --username=root --owner=root go-simple-bank

drop-db:
	docker exec -it go-simple-bank-db dropdb go-simple-bank

migrate-up:
	migrate -path migrations -database "postgresql://root:secret@localhost:5433/go-simple-bank?sslmode=disable" -verbose up
migrate-down:
	migrate -path migrations -database "postgresql://root:secret@localhost:5433/go-simple-bank?sslmode=disable" -verbose down

# Create migration file
cm:
	@migrate create -ext sql -dir migrations -seq $(a)

sqlc:
	@cd "internal/infrastructure/database/"; sqlc generate

test:
	go test -v -cover ./...

.PNONY: init build run clean local-git git-init postgres create-db drop-db migrate-up migrate-down sqlc