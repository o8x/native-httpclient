V              = ""
LATEST_VERSION = $(shell cat native_http_client.go | grep Native_Http_Client | awk '{print $$4}' | sed 's/\//\\\//g')

release:
	sed -i "" 's/$(LATEST_VERSION)/"Native_Http_Client\/$(V)"/g' native_http_client.go
	git commit -am "release v$(V)"
	git push
	gh release create "v$(V)"
