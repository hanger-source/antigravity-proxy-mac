.PHONY: build build-helper install clean run

APP_NAME = AntigravityProxy
BUILD_DIR = build

build: build-helper
	go build -o $(BUILD_DIR)/$(APP_NAME).app/Contents/MacOS/antigravity-proxy .
	cp helper/antigravity-proxy-helper $(BUILD_DIR)/$(APP_NAME).app/Contents/Resources/
	cp helper/com.antigravity-proxy.helper.plist $(BUILD_DIR)/$(APP_NAME).app/Contents/Resources/

build-helper:
	cd helper && go build -o antigravity-proxy-helper .

install: build
	cp -r $(BUILD_DIR)/$(APP_NAME).app /Applications/$(APP_NAME).app

clean:
	rm -rf $(BUILD_DIR)/$(APP_NAME).app/Contents/MacOS/antigravity-proxy
	rm -f helper/antigravity-proxy-helper

run: build-helper
	go run .
