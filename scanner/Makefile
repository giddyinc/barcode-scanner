.PHONY: doc local

default: local

doc:
	godoc -http=:8765

local:
	go build -o scale
	while true; do scale && sleep 1; done
