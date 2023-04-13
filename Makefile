make clean-build:
	rm dist
	rm coverage
	rm node_modules

setup-frontend:
	yarn install

build-frontend:
	yarn build

install-mage:
	go install github.com/magefile/mage

setup-backend:
	go get -u github.com/grafana/grafana-plugin-sdk-go
	go mod tidy

build-backend:
	# build plugin's backend
	mage -v
	
	# required by plugincheck
	chmod 0775 dist/gpx_websocket_* 

sign-plugin:
	npm run sign

validate-plugin:
	cp -r dist golioth-websocket-datasource
	zip golioth-websocket-datasource golioth-websocket-datasource -r
	npx -y @grafana/plugin-validator@latest ./golioth-websocket-datasource.zip || true
	rm -r golioth-websocket-datasource
	rm golioth-websocket-datasource.zip
