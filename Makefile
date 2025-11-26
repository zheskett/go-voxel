.PHONY: all
all: build

.PHONY: build
build:
	go build -o go-voxel cmd/go-voxel/go-voxel.go

.PHONY: run
run: build
	./go-voxel

.PHONY: clean
clean:
	rm -rf go-voxel
