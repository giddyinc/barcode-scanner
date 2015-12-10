.PHONY: doc local

default: local

doc:
	godoc -http=:8765

local:
	go build -o barcode
	./barcode
