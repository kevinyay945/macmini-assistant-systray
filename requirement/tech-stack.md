OS: Mac
Language: Golang


- a orchestra to call the tool register at the copilot-sdk
- lib:
	- downie (with deep link)
	- google drive upload
- trigger method
	- line-bot
		- reply mode(bot only reply after user send the message)
	- discord
		- reply mode
		- send server status
- feature
	- set the download folder(default is in the tmp folder)
	- register on systray
	- auto start after startup
	- auto download from github release
	- llm timeout: 10 mins
	- support tool extension
		- we can trigger cli tool create new tool on cloud to support more feature
	- chatbot
		- line/discord
			- directly reply from llm
			- call tool
	- status panel channel (only discord)
		- send every tool calling from copilot-sdk
		- error message report
- tool
	- youtube download
		- input:
			- youtube-link
			- file-format
			- resolution
		- output:
			-  file-path(abs path)
			- file-size
	- google drive upload and generate share link
		- input:
			- file-path
			- upload-name
			- timeout
		- output:
			-  share link


Lib:
- github.com/getlantern/systray
- github.com/gin-gonic/gin
- github.com/spf13/cobra.git
- github.com/line/line-bot-sdk-go/v8/linebot
- github.com/inconshreveable/go-update
