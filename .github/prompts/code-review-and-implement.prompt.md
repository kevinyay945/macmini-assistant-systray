---
model: Claude Opus 4.5 (copilot)
---

@workspace code review by run the command ```git --no-pager diff --name-only main...HEAD --no-color ``` , if there are any suggest, give me the link to reference to the target file and target line in this #codebase
this reference link should use relative path
after that, implement the changes for me based on your suggestions and test and lint the code to make sure everything works fine and update the spec files if needed
after that , repeat the code review again to make sure everything is perfect
