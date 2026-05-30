.PHONY: build build-helper install clean run

APP_NAME = Funnel
BUILD_DIR = build

build: build-helper
	go build -o $(BUILD_DIR)/$(APP_NAME).app/Contents/MacOS/funnel .
	cp helper/funnel-helper $(BUILD_DIR)/$(APP_NAME).app/Contents/Resources/
	cp helper/com.funnel.helper.plist $(BUILD_DIR)/$(APP_NAME).app/Contents/Resources/

build-helper:
	cd helper && go build -o funnel-helper .

install: build
	cp -r $(BUILD_DIR)/$(APP_NAME).app /Applications/$(APP_NAME).app

clean:
	rm -rf $(BUILD_DIR)/$(APP_NAME).app/Contents/MacOS/funnel
	rm -f helper/funnel-helper

run: build-helper
	go run .
