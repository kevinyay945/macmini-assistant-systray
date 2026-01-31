# Kitchen Sink Example

A comprehensive LINE bot example that demonstrates handling multiple event types and message types.

## Features
- Handles text, image, video, audio, file, location, and sticker messages
- Processes follow/unfollow events
- Manages join/leave events for groups and rooms
- Handles postback events
- Demonstrates file download and storage
- Shows template messages and flex messages
- Includes rich menu operations

## Code

See the full implementation in [kitchensink.go](./kitchensink.go)

## Architecture

```go
type KitchenSink struct {
    channelSecret string
    bot           *messaging_api.MessagingApiAPI
    blob          *messaging_api.MessagingApiBlobAPI
    appBaseURL    string
    downloadDir   string
}
```

The bot maintains:
- **bot**: Main messaging API client
- **blob**: Blob API client for downloading media
- **appBaseURL**: Base URL for serving content back to LINE
- **downloadDir**: Local directory for storing downloaded files

## Key Concepts Demonstrated

### 1. Multiple API Clients

```go
// Messaging API for sending messages
bot, err := messaging_api.NewMessagingApiAPI(channelToken)

// Blob API for downloading content
blob, err := messaging_api.NewMessagingApiBlobAPI(channelToken)
```

### 2. Comprehensive Event Handling

The kitchen sink handles all major event types:

**Message Events:**
- Text messages
- Image messages
- Video messages
- Audio messages
- File messages
- Location messages
- Sticker messages

**Relationship Events:**
- Follow (user adds bot as friend)
- Unfollow (user blocks bot)
- Join (bot added to group/room)
- Leave (bot removed from group/room)

**Interactive Events:**
- Postback (from buttons/quick replies)
- Beacon (proximity-based events)

### 3. Content Download Pattern

```go
func (app *KitchenSink) handleImage(replyToken, messageID string) error {
    // Download image content
    content, resp, err := app.blob.GetMessageContentWithHttpInfo(messageID)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Save to file
    filepath := filepath.Join(app.downloadDir, messageID+".jpg")
    file, err := os.Create(filepath)
    if err != nil {
        return err
    }
    defer file.Close()

    if _, err := io.Copy(file, resp.Body); err != nil {
        return err
    }

    // Reply with downloaded content info
    return app.replyText(replyToken, "Image downloaded")
}
```

### 4. Template Messages

**Buttons Template:**
```go
messaging_api.TemplateMessage{
    AltText: "Buttons template",
    Template: &messaging_api.ButtonsTemplate{
        ThumbnailImageUrl: app.appBaseURL + "/static/buttons/image.jpg",
        Title:             "My Button",
        Text:              "Please select",
        Actions: []messaging_api.ActionInterface{
            &messaging_api.PostbackAction{
                Label: "Postback",
                Data:  "action=buy&itemid=1",
            },
            &messaging_api.MessageAction{
                Label: "Message",
                Text:  "message text",
            },
            &messaging_api.URIAction{
                Label: "URI",
                Uri:   "https://example.com",
            },
        },
    },
}
```

**Confirm Template:**
```go
messaging_api.TemplateMessage{
    AltText: "Confirm template",
    Template: &messaging_api.ConfirmTemplate{
        Text: "Do you want to proceed?",
        Actions: []messaging_api.ActionInterface{
            &messaging_api.MessageAction{
                Label: "Yes",
                Text:  "yes",
            },
            &messaging_api.MessageAction{
                Label: "No",
                Text:  "no",
            },
        },
    },
}
```

**Carousel Template:**
```go
messaging_api.TemplateMessage{
    AltText: "Carousel template",
    Template: &messaging_api.CarouselTemplate{
        Columns: []messaging_api.CarouselColumn{
            {
                ThumbnailImageUrl: app.appBaseURL + "/static/item1.jpg",
                Title:             "Item 1",
                Text:              "Description 1",
                Actions: []messaging_api.ActionInterface{
                    &messaging_api.PostbackAction{
                        Label: "Buy",
                        Data:  "action=buy&itemid=1",
                    },
                },
            },
            // More columns...
        },
    },
}
```

### 5. Flex Messages

Flex messages provide highly customizable layouts:

```go
messaging_api.FlexMessage{
    AltText: "Flex message",
    Contents: &messaging_api.FlexBubble{
        Type: "bubble",
        Hero: &messaging_api.FlexImage{
            Type: "image",
            Url:  app.appBaseURL + "/static/hero.jpg",
            Size: "full",
        },
        Body: &messaging_api.FlexBox{
            Type:   "vertical",
            Layout: "vertical",
            Contents: []messaging_api.FlexComponentInterface{
                &messaging_api.FlexText{
                    Type:   "text",
                    Text:   "Title",
                    Size:   "xl",
                    Weight: "bold",
                },
                &messaging_api.FlexText{
                    Type:  "text",
                    Text:  "Description",
                    Size:  "sm",
                    Color: "#999999",
                },
            },
        },
        Footer: &messaging_api.FlexBox{
            Type:   "vertical",
            Layout: "vertical",
            Contents: []messaging_api.FlexComponentInterface{
                &messaging_api.FlexButton{
                    Type:   "button",
                    Action: &messaging_api.URIAction{
                        Label: "View More",
                        Uri:   "https://example.com",
                    },
                },
            },
        },
    },
}
```

### 6. Quick Replies

```go
bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
    ReplyToken: replyToken,
    Messages: []messaging_api.MessageInterface{
        messaging_api.TextMessage{
            Text: "Select an option:",
            QuickReply: &messaging_api.QuickReply{
                Items: []messaging_api.QuickReplyItem{
                    {
                        Type: "action",
                        Action: &messaging_api.MessageAction{
                            Label: "Option 1",
                            Text:  "option1",
                        },
                    },
                    {
                        Type: "action",
                        Action: &messaging_api.MessageAction{
                            Label: "Option 2",
                            Text:  "option2",
                        },
                    },
                },
            },
        },
    },
})
```

### 7. User Profile Information

```go
func (app *KitchenSink) getUserProfile(userID string) (*messaging_api.UserProfileResponse, error) {
    profile, _, err := app.bot.GetProfileWithHttpInfo(userID)
    if err != nil {
        return nil, err
    }
    
    // profile contains:
    // - DisplayName
    // - UserId
    // - PictureUrl
    // - StatusMessage
    
    return profile, nil
}
```

### 8. Group/Room Information

```go
// Get group summary
groupID := event.Source.GroupId
summary, _, err := app.bot.GetGroupSummaryWithHttpInfo(groupID)

// Get room members count
roomID := event.Source.RoomId
count, _, err := app.bot.GetRoomMembersCountWithHttpInfo(roomID)
```

### 9. Leave Group/Room

```go
// Leave a group
func (app *KitchenSink) leaveGroup(groupID string) error {
    _, err := app.bot.LeaveGroup(groupID)
    return err
}

// Leave a room
func (app *KitchenSink) leaveRoom(roomID string) error {
    _, err := app.bot.LeaveRoom(roomID)
    return err
}
```

### 10. Static File Serving

The kitchen sink serves static files for images used in messages:

```go
// Serve static files
staticFileServer := http.FileServer(http.Dir(staticDir))
http.HandleFunc("/static/", http.StripPrefix("/static/", staticFileServer).ServeHTTP)

// Serve downloaded files
downloadedFileServer := http.FileServer(http.Dir(app.downloadDir))
http.HandleFunc("/downloaded/", http.StripPrefix("/downloaded/", downloadedFileServer).ServeHTTP)
```

## Running the Example

```bash
# Set environment variables
export LINE_CHANNEL_SECRET=your_channel_secret
export LINE_CHANNEL_TOKEN=your_channel_token
export APP_BASE_URL=https://your-domain.com  # Your public HTTPS URL
export PORT=5000  # Optional

# Run the bot
go run kitchensink.go
```

## Directory Structure

```
kitchensink/
├── server.go           # Main application
├── static/             # Static files for messages
│   ├── buttons/
│   ├── carousel/
│   └── ...
└── line-bot/          # Downloaded content (created at runtime)
```

## Use Cases

This example is perfect for:
- **Learning**: Demonstrates all major features of the SDK
- **Reference**: Shows best practices for each message type
- **Testing**: Useful for testing different message formats
- **Prototyping**: Base for building more complex bots

## Production Considerations

1. **Error Handling**: Add comprehensive error handling
2. **Logging**: Implement structured logging
3. **Database**: Store user data and conversation state
4. **Queue**: Use message queue for async processing
5. **Monitoring**: Add health checks and metrics
6. **Security**: Validate all inputs, rate limiting
7. **HTTPS**: Use proper TLS certificates

## Related Examples

- **Echo Bot**: Simpler example focusing on basic echo functionality
- **Echo Bot Handler**: Handler-based implementation with automatic signature verification
- **Delivery Helper**: Demonstrates delivery status checking
- **Insight Helper**: Shows analytics and insights API usage
- **Rich Menu Helper**: Rich menu creation and management
