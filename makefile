GitTag = $(shell git describe --abbrev=0 --tags)

release:
	goreleaser release --snapshot --rm-dist
	echo $(GitTag) > ./dist/version