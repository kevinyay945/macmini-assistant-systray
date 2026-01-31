---
tools: ["search/codebase", "search/changes", "read/problems", "execute/getTerminalOutput", "execute/runInTerminal", "read/terminalLastCommand", "read/terminalSelection", "search", "search/searchResults", "search/usages",]
model: Claude Opus 4.5 (copilot)
---

@workspace code review by run the command ```git --no-pager diff main...HEAD --no-color ``` , if there are any suggest, give me the link to reference to the target file and target line in this #codebase
this reference link should use relative path
don't edit my file, only suggestion and example code
use zh-tw to reply
for your suggestion, please give me the reason why need to change, and if possible, give me an example code block to show how to change
