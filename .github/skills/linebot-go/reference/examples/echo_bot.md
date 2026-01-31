# Echo Bot Example

This is the simplest LINE bot implementation that echoes back any text message received.

## Features
- Receives webhook events from LINE
- Echoes back text messages
- Handles sticker messages
- Demonstrates basic message reply pattern

## Code

See the full implementation in [echo_bot.go](./echo_bot.go)

## Key Concepts Demonstrated

### 1. Webhook Request Parsing
```go
cb, err := webhook.ParseRequest(channelSecret, req)
if err != nil {
    if errors.Is(err, webhook.ErrInvalidSignature) {
        w.WriteHeader(400)
    } else {
        w.WriteHeader(500)
    }
    return
}
```

### 2. Event Type Switching
```go
for _, event := range cb.Events {
    switch e := event.(type) {
    case webhook.MessageEvent:
        // Handle message events
    default:
        log.Printf("Unsupported message: %T\n", event)
    }
}
```

### 3. Message Type Handling
```go
switch message := e.Message.(type) {
case webhook.TextMessageContent:
    // Echo the text back
    bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
        ReplyToken: e.ReplyToken,
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{
                Text: message.Text,
            },
        },
    })

case webhook.StickerMessageContent:
    // Reply with sticker info
    replyMessage := fmt.Sprintf(
        "sticker id is %s, stickerResourceType is %s", 
        message.StickerId, 
        message.StickerResourceType)
    bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
        ReplyToken: e.ReplyToken,
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{
                Text: replyMessage,
            },
        },
    })
}
```

### 4. Text Message with Emoji
```go
messaging_api.TextMessageV2{
    Text: "Hello! {smile}",
    Substitution: map[string]messaging_api.SubstitutionObjectInterface{
        "smile": &messaging_api.EmojiSubstitutionObject{
            ProductId: "5ac1bfd5040ab15980c9b435",
            EmojiId:   "002",
        },
    },
}
```

## Running the Example

```bash
# Set environment variables
export LINE_CHANNEL_SECRET=your_channel_secret
export LINE_CHANNEL_TOKEN=your_channel_token
export PORT=5000  # Optional

# Run the bot
go run echo_bot.go
```

## Testing

1. Start the server
2. Use ngrok to expose your local server:
   ```bash
   ngrok http 5000
   ```
3. Set the ngrok HTTPS URL as your webhook URL in LINE Developers Console
4. Send a message to your bot on LINE
5. The bot should echo your message back

## Production Deployment

⚠️ **Important**: This example uses HTTP (`http.ListenAndServe`). For production:
- Use HTTPS with `http.ListenAndServeTLS`
- Or deploy behind a reverse proxy (nginx, Caddy)
- Or use a cloud platform with automatic HTTPS (Heroku, Cloud Run, etc.)

LINE requires HTTPS for webhook URLs in production.
