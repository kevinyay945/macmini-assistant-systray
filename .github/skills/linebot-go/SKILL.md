---
name: linebot-go
description: Expert guidance for using LINE Messaging API SDK with Go, including webhook handling, message types, and bot implementation patterns.
---

# LINE Messaging API SDK for Go - Skill Guide

This skill provides comprehensive guidance for building LINE bots using the official LINE Messaging API SDK for Go.

## Reference Documentation Locations

### Core SDK Documentation
- **Official LINE SDK Repository**: `line-bot-sdk-go/`
- **Main README**: [./reference/README.md](./reference/README.md) - Complete API reference and installation
- **Official LINE Docs**: https://developers.line.biz/en/docs/messaging-api/overview/

### Example Implementations
- **Echo Bot**: [./reference/examples/echo_bot.md](./reference/examples/echo_bot.md) - Simple message echo bot
- **Kitchen Sink**: [./reference/examples/kitchensink.md](./reference/examples/kitchensink.md) - Comprehensive example with multiple event types
- **Echo Bot Handler**: [./reference/examples/echo_bot_handler.md](./reference/examples/echo_bot_handler.md) - Handler-based implementation

## Installation

```bash
go get -u github.com/line/line-bot-sdk-go/v8/linebot
```

## Required Environment Variables

```bash
LINE_CHANNEL_SECRET=your_channel_secret
LINE_CHANNEL_TOKEN=your_channel_access_token
PORT=5000  # Optional, defaults to 5000
```

## Core Concepts

### 1. Client Initialization

```go
import (
    "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
    "github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func main() {
    channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
    channelToken := os.Getenv("LINE_CHANNEL_TOKEN")
    
    // Create messaging API client
    bot, err := messaging_api.NewMessagingApiAPI(channelToken)
    if err != nil {
        log.Fatal(err)
    }
    
    // For blob operations (images, videos, audio)
    blob, err := messaging_api.NewMessagingApiBlobAPI(channelToken)
    if err != nil {
        log.Fatal(err)
    }
}
```

**Client Configuration Options:**
- `WithHTTPClient(client)`: Use custom HTTP client
- `WithEndpoint(endpoint)`: Use custom API endpoint
- `WithBlobHTTPClient(client)`: Custom HTTP client for blob operations
- `WithBlobEndpoint(endpoint)`: Custom blob API endpoint

### 2. Webhook Event Handling

```go
http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
    // Parse incoming webhook request
    cb, err := webhook.ParseRequest(channelSecret, req)
    if err != nil {
        if errors.Is(err, webhook.ErrInvalidSignature) {
            w.WriteHeader(400)
        } else {
            w.WriteHeader(500)
        }
        return
    }
    
    // Process each event
    for _, event := range cb.Events {
        switch e := event.(type) {
        case webhook.MessageEvent:
            handleMessageEvent(bot, e)
        case webhook.FollowEvent:
            handleFollowEvent(bot, e)
        case webhook.UnfollowEvent:
            handleUnfollowEvent(bot, e)
        case webhook.JoinEvent:
            handleJoinEvent(bot, e)
        case webhook.LeaveEvent:
            handleLeaveEvent(bot, e)
        case webhook.PostbackEvent:
            handlePostbackEvent(bot, e)
        case webhook.BeaconEvent:
            handleBeaconEvent(bot, e)
        default:
            log.Printf("Unknown event type: %T\n", event)
        }
    }
})
```

### 3. Message Event Types

```go
func handleMessageEvent(bot *messaging_api.MessagingApiAPI, e webhook.MessageEvent) {
    switch message := e.Message.(type) {
    case webhook.TextMessageContent:
        // Handle text message
        log.Printf("Text: %s", message.Text)
        
    case webhook.ImageMessageContent:
        // Handle image message
        log.Printf("Image ID: %s", message.Id)
        
    case webhook.VideoMessageContent:
        // Handle video message
        log.Printf("Video ID: %s", message.Id)
        
    case webhook.AudioMessageContent:
        // Handle audio message
        log.Printf("Audio ID: %s, Duration: %d", message.Id, message.Duration)
        
    case webhook.FileMessageContent:
        // Handle file message
        log.Printf("File: %s (%d bytes)", message.FileName, message.FileSize)
        
    case webhook.LocationMessageContent:
        // Handle location message
        log.Printf("Location: %s (%f, %f)", message.Address, message.Latitude, message.Longitude)
        
    case webhook.StickerMessageContent:
        // Handle sticker message
        log.Printf("Sticker: %s (Type: %s)", message.StickerId, message.StickerResourceType)
        
    default:
        log.Printf("Unknown message type: %T", message)
    }
}
```

### 4. Sending Messages

#### Reply to Messages (using Reply Token)

```go
// Simple text reply
_, err = bot.ReplyMessage(
    &messaging_api.ReplyMessageRequest{
        ReplyToken: e.ReplyToken,
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{
                Text: "Hello, World!",
            },
        },
    },
)
```

#### Push Messages (using User/Group/Room ID)

```go
// Push message to user
_, err = bot.PushMessage(
    &messaging_api.PushMessageRequest{
        To: "U1234567890abcdef1234567890abcdef", // User ID
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{
                Text: "Hello!",
            },
        },
    },
    "", // x-line-retry-key (optional)
)
```

#### Multicast Messages (to multiple users)

```go
_, err = bot.Multicast(
    &messaging_api.MulticastRequest{
        To: []string{
            "U1234567890abcdef1234567890abcdef",
            "U2234567890abcdef1234567890abcdef",
        },
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{
                Text: "Broadcast message",
            },
        },
    },
    "", // x-line-retry-key (optional)
)
```

### 5. Message Types

#### Text Message with Emojis

```go
messaging_api.TextMessageV2{
    Text: "Hello! {smile} {heart}",
    Substitution: map[string]messaging_api.SubstitutionObjectInterface{
        "smile": &messaging_api.EmojiSubstitutionObject{
            ProductId: "5ac1bfd5040ab15980c9b435",
            EmojiId:   "002",
        },
        "heart": &messaging_api.EmojiSubstitutionObject{
            ProductId: "5ac1bfd5040ab15980c9b435",
            EmojiId:   "001",
        },
    },
}
```

#### Image Message

```go
messaging_api.ImageMessage{
    OriginalContentUrl: "https://example.com/original.jpg",
    PreviewImageUrl:    "https://example.com/preview.jpg",
}
```

#### Video Message

```go
messaging_api.VideoMessage{
    OriginalContentUrl: "https://example.com/video.mp4",
    PreviewImageUrl:    "https://example.com/preview.jpg",
}
```

#### Audio Message

```go
messaging_api.AudioMessage{
    OriginalContentUrl: "https://example.com/audio.m4a",
    Duration:           60000, // milliseconds
}
```

#### Location Message

```go
messaging_api.LocationMessage{
    Title:     "My Location",
    Address:   "Tokyo, Japan",
    Latitude:  35.6812,
    Longitude: 139.7671,
}
```

#### Sticker Message

```go
messaging_api.StickerMessage{
    PackageId: "446",
    StickerId: "1988",
}
```

#### Template Messages

**Buttons Template:**
```go
messaging_api.TemplateMessage{
    AltText: "This is a buttons template",
    Template: &messaging_api.ButtonsTemplate{
        ThumbnailImageUrl: "https://example.com/image.jpg",
        Title:             "Menu",
        Text:              "Please select",
        Actions: []messaging_api.ActionInterface{
            &messaging_api.PostbackAction{
                Label: "Buy",
                Data:  "action=buy&itemid=123",
            },
            &messaging_api.MessageAction{
                Label: "Say hello",
                Text:  "hello",
            },
            &messaging_api.URIAction{
                Label: "View detail",
                Uri:   "https://example.com",
            },
        },
    },
}
```

**Confirm Template:**
```go
messaging_api.TemplateMessage{
    AltText: "This is a confirm template",
    Template: &messaging_api.ConfirmTemplate{
        Text: "Are you sure?",
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
    AltText: "This is a carousel template",
    Template: &messaging_api.CarouselTemplate{
        Columns: []messaging_api.CarouselColumn{
            {
                ThumbnailImageUrl: "https://example.com/item1.jpg",
                Title:             "Item 1",
                Text:              "Description 1",
                Actions: []messaging_api.ActionInterface{
                    &messaging_api.PostbackAction{
                        Label: "Buy",
                        Data:  "item=1",
                    },
                },
            },
            {
                ThumbnailImageUrl: "https://example.com/item2.jpg",
                Title:             "Item 2",
                Text:              "Description 2",
                Actions: []messaging_api.ActionInterface{
                    &messaging_api.PostbackAction{
                        Label: "Buy",
                        Data:  "item=2",
                    },
                },
            },
        },
    },
}
```

#### Flex Message

```go
messaging_api.FlexMessage{
    AltText: "This is a flex message",
    Contents: &messaging_api.FlexBubble{
        Type: "bubble",
        Body: &messaging_api.FlexBox{
            Type:   "vertical",
            Layout: "vertical",
            Contents: []messaging_api.FlexComponentInterface{
                &messaging_api.FlexText{
                    Type: "text",
                    Text: "Hello, Flex!",
                    Size: "xl",
                    Weight: "bold",
                },
            },
        },
    },
}
```

### 6. Getting User Information

```go
// Get user profile from User ID
userID := event.Source.UserId
profile, resp, err := bot.GetProfileWithHttpInfo(userID)
if err != nil {
    log.Printf("Error getting profile: %v", err)
} else {
    log.Printf("Display Name: %s", profile.DisplayName)
    log.Printf("User ID: %s", profile.UserId)
    log.Printf("Picture URL: %s", profile.PictureUrl)
    log.Printf("Status Message: %s", profile.StatusMessage)
}
```

### 7. Response Headers and Error Handling

#### Get Response Headers

```go
resp, _, err := bot.ReplyMessageWithHttpInfo(
    &messaging_api.ReplyMessageRequest{
        ReplyToken: replyToken,
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{
                Text: "Hello, world",
            },
        },
    },
)

if err == nil {
    log.Printf("Status: %d", resp.StatusCode)
    log.Printf("Request ID: %s", resp.Header.Get("x-line-request-id"))
}
```

#### Handle Error Responses

```go
resp, _, err := bot.ReplyMessageWithHttpInfo(request)
if err != nil && resp.StatusCode >= 400 && resp.StatusCode < 500 {
    decoder := json.NewDecoder(resp.Body)
    errorResponse := &messaging_api.ErrorResponse{}
    if err := decoder.Decode(&errorResponse); err != nil {
        log.Printf("Failed to decode error: %v", err)
    } else {
        log.Printf("Error: %s (Request ID: %s)", 
            errorResponse.Message, 
            resp.Header.Get("x-line-request-id"))
    }
}
```

### 8. Downloading Content (Images, Videos, Audio)

```go
// Download image/video/audio content
messageID := "message-id-from-webhook"

content, resp, err := blob.GetMessageContentWithHttpInfo(messageID)
if err != nil {
    log.Printf("Error downloading content: %v", err)
    return
}
defer resp.Body.Close()

// Save to file
file, err := os.Create("downloaded-content")
if err != nil {
    log.Printf("Error creating file: %v", err)
    return
}
defer file.Close()

_, err = io.Copy(file, resp.Body)
if err != nil {
    log.Printf("Error saving file: %v", err)
}
```

### 9. Source Types (User, Group, Room)

```go
// Get source information from event
switch source := event.Source.(type) {
case webhook.UserSource:
    userID := source.UserId
    log.Printf("Message from user: %s", userID)
    
case webhook.GroupSource:
    groupID := source.GroupId
    userID := source.UserId
    log.Printf("Message from user %s in group %s", userID, groupID)
    
case webhook.RoomSource:
    roomID := source.RoomId
    userID := source.UserId
    log.Printf("Message from user %s in room %s", userID, roomID)
}
```

### 10. Postback Events

```go
func handlePostbackEvent(bot *messaging_api.MessagingApiAPI, e webhook.PostbackEvent) {
    data := e.Postback.Data
    log.Printf("Postback data: %s", data)
    
    // Parse postback data (e.g., "action=buy&itemid=123")
    params, _ := url.ParseQuery(data)
    action := params.Get("action")
    itemID := params.Get("itemid")
    
    // Respond to postback
    bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
        ReplyToken: e.ReplyToken,
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{
                Text: fmt.Sprintf("Action: %s, Item: %s", action, itemID),
            },
        },
    })
}
```

## Common Patterns

### Pattern 1: Echo Bot (Simplest)

```go
for _, event := range cb.Events {
    switch e := event.(type) {
    case webhook.MessageEvent:
        switch message := e.Message.(type) {
        case webhook.TextMessageContent:
            bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
                ReplyToken: e.ReplyToken,
                Messages: []messaging_api.MessageInterface{
                    messaging_api.TextMessage{
                        Text: message.Text,
                    },
                },
            })
        }
    }
}
```

### Pattern 2: Command-Based Bot

```go
case webhook.TextMessageContent:
    text := message.Text
    var replyText string
    
    switch {
    case strings.HasPrefix(text, "/help"):
        replyText = "Available commands: /help, /about, /weather"
    case strings.HasPrefix(text, "/about"):
        replyText = "I'm a LINE bot built with Go!"
    case strings.HasPrefix(text, "/weather"):
        replyText = "Weather info goes here..."
    default:
        replyText = "Unknown command. Type /help for available commands."
    }
    
    bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
        ReplyToken: e.ReplyToken,
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{Text: replyText},
        },
    })
```

### Pattern 3: Multiple Message Reply

```go
bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
    ReplyToken: e.ReplyToken,
    Messages: []messaging_api.MessageInterface{
        messaging_api.TextMessage{
            Text: "First message",
        },
        messaging_api.TextMessage{
            Text: "Second message",
        },
        messaging_api.ImageMessage{
            OriginalContentUrl: "https://example.com/image.jpg",
            PreviewImageUrl:    "https://example.com/preview.jpg",
        },
    },
})
```

## Best Practices

1. **Always validate signatures** - Use `webhook.ParseRequest()` to verify requests are from LINE
2. **Handle errors gracefully** - Check all error returns and log appropriately
3. **Use HTTPS in production** - LINE requires HTTPS for webhook endpoints
4. **Return 200 OK quickly** - Process events asynchronously if needed
5. **Store channel credentials securely** - Use environment variables, never hardcode
6. **Respect rate limits** - LINE has API rate limits, implement retry logic
7. **Log request IDs** - Use `x-line-request-id` header for debugging
8. **Test with ngrok locally** - Use ngrok to expose local server for testing

## Quick Start Template

```go
package main

import (
    "errors"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/line/line-bot-sdk-go/v8/linebot/messaging_api"
    "github.com/line/line-bot-sdk-go/v8/linebot/webhook"
)

func main() {
    channelSecret := os.Getenv("LINE_CHANNEL_SECRET")
    channelToken := os.Getenv("LINE_CHANNEL_TOKEN")
    
    bot, err := messaging_api.NewMessagingApiAPI(channelToken)
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
        cb, err := webhook.ParseRequest(channelSecret, req)
        if err != nil {
            if errors.Is(err, webhook.ErrInvalidSignature) {
                w.WriteHeader(400)
            } else {
                w.WriteHeader(500)
            }
            return
        }

        for _, event := range cb.Events {
            switch e := event.(type) {
            case webhook.MessageEvent:
                switch message := e.Message.(type) {
                case webhook.TextMessageContent:
                    bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
                        ReplyToken: e.ReplyToken,
                        Messages: []messaging_api.MessageInterface{
                            messaging_api.TextMessage{
                                Text: message.Text,
                            },
                        },
                    })
                }
            }
        }
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "5000"
    }
    
    fmt.Printf("Server running on :%s\n", port)
    if err := http.ListenAndServe(":"+port, nil); err != nil {
        log.Fatal(err)
    }
}
```

## Testing Locally

```bash
# 1. Install ngrok
brew install ngrok

# 2. Start your bot
export LINE_CHANNEL_SECRET=your_secret
export LINE_CHANNEL_TOKEN=your_token
go run main.go

# 3. In another terminal, expose local server
ngrok http 5000

# 4. Copy the HTTPS URL from ngrok and set it as your webhook URL in LINE Developers Console
# Example: https://abc123.ngrok.io/callback
```

## Resources

- **LINE Developers**: https://developers.line.biz/
- **Messaging API Reference**: https://developers.line.biz/en/reference/messaging-api/
- **SDK Repository**: https://github.com/line/line-bot-sdk-go
- **SDK Documentation**: https://pkg.go.dev/github.com/line/line-bot-sdk-go/v8/linebot
- **Flex Message Simulator**: https://developers.line.biz/flex-simulator/

## Troubleshooting

### Signature Validation Failed
- Verify `LINE_CHANNEL_SECRET` is correct
- Check request body hasn't been modified
- Ensure you're using `webhook.ParseRequest()` correctly

### Messages Not Sending
- Check `LINE_CHANNEL_TOKEN` is valid
- Verify webhook URL is HTTPS in production
- Check API response for error details using `*WithHttpInfo` methods

### Reply Token Invalid
- Reply tokens can only be used once
- Reply tokens expire after a certain time
- Use Push API for delayed responses

### Bot Not Responding
- Check server logs for errors
- Verify webhook URL in LINE Developers Console
- Test webhook endpoint is accessible publicly
- Ensure server returns 200 OK within timeout
