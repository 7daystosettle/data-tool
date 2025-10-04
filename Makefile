run: build
	cd bin && ./data-tool "items.xml" "items.kdl"
	cd bin && ./data-tool "items.kdl" "items_out.xml"

build:
	go build -o bin/data-tool main.go