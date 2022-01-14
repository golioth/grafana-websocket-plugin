setup:
	go get -u github.com/grafana/plugin-validator/cmd/plugincheck

validate-plugin:
	cp -r dist golioth-websocket-datasource
	zip golioth-websocket-datasource golioth-websocket-datasource -r
	plugincheck ./golioth-websocket-datasource.zip || true
	rm -r golioth-websocket-datasource
	rm golioth-websocket-datasource.zip