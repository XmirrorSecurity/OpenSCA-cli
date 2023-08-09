GitTag = $(shell git describe --abbrev=0 --tags)

release:
	goreleaser release --snapshot --clean
	echo $(GitTag) > ./dist/version