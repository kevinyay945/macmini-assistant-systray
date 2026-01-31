# LINE Bot SDK Go Examples

This directory contains example implementations demonstrating various features of the LINE Messaging API SDK for Go.

## Available Examples

### 1. Echo Bot ([echo_bot.md](./echo_bot.md))
**Difficulty**: Beginner  
**File**: [echo_bot.go](./echo_bot.go)

A simple bot that echoes back text messages and handles stickers. Perfect for:
- Getting started with LINE bots
- Understanding webhook basics
- Learning message reply patterns

**Key Features:**
- Basic webhook handling
- Text message echo
- Sticker message handling
- Text messages with emojis

---

### 2. Kitchen Sink ([kitchensink.md](./kitchensink.md))
**Difficulty**: Advanced  
**File**: [kitchensink.go](./kitchensink.go)

A comprehensive example showing all major SDK features. Perfect for:
- Learning advanced bot patterns
- Reference implementation for all message types
- Understanding event handling

**Key Features:**
- All message types (text, image, video, audio, file, location, sticker)
- Template messages (buttons, confirm, carousel)
- Flex messages
- Quick replies
- Content download
- User profile retrieval
- Group/room management
- Static file serving

---

### 3. Echo Bot Handler (Available in SDK)
**Difficulty**: Beginner  
**Path**: `line-bot-sdk-go/examples/echo_bot_handler/`

Handler-based implementation with automatic signature verification.

---

## Quick Start

### Prerequisites

1. **LINE Developer Account**: https://developers.line.biz/
2. **Create a Channel**: Create a Messaging API channel
3. **Get Credentials**:
   - Channel Secret
   - Channel Access Token

### Setup

1. **Install the SDK**:
   ```bash
   go get -u github.com/line/line-bot-sdk-go/v8/linebot
   ```

2. **Set Environment Variables**:
   ```bash
   export LINE_CHANNEL_SECRET=your_channel_secret_here
   export LINE_CHANNEL_TOKEN=your_channel_access_token_here
   export PORT=5000
   ```

3. **Run an Example**:
   ```bash
   # Echo Bot
   go run echo_bot.go

   # Kitchen Sink (requires APP_BASE_URL)
   export APP_BASE_URL=https://your-domain.com
   go run kitchensink.go
   ```

### Local Testing with ngrok

LINE requires HTTPS for webhooks. Use ngrok for local development:

```bash
# 1. Install ngrok
brew install ngrok

# 2. Start your bot
go run echo_bot.go

# 3. In another terminal, expose your local server
ngrok http 5000

# 4. Copy the HTTPS URL (e.g., https://abc123.ngrok.io)
# 5. Set webhook URL in LINE Developers Console:
#    https://abc123.ngrok.io/callback
```

## Example Comparison

| Feature | Echo Bot | Kitchen Sink |
|---------|----------|--------------|
| Text Messages | ✅ | ✅ |
| Image Messages | ❌ | ✅ |
| Video Messages | ❌ | ✅ |
| Audio Messages | ❌ | ✅ |
| File Messages | ❌ | ✅ |
| Location Messages | ❌ | ✅ |
| Sticker Messages | ✅ | ✅ |
| Template Messages | ❌ | ✅ |
| Flex Messages | ❌ | ✅ |
| Quick Replies | ❌ | ✅ |
| Content Download | ❌ | ✅ |
| User Profile | ❌ | ✅ |
| Group/Room Events | ❌ | ✅ |
| Postback Events | ❌ | ✅ |
| Static File Serving | ❌ | ✅ |

## Learning Path

### 1. Start with Echo Bot
- Understand webhook mechanics
- Learn basic message handling
- Get comfortable with the SDK structure

### 2. Explore Kitchen Sink
- Study advanced message types
- Learn template and flex messages
- Understand event handling patterns

### 3. Build Your Own
- Start with Echo Bot as a template
- Add features from Kitchen Sink as needed
- Implement your custom logic

## Common Patterns

### Pattern: Simple Echo
```go
case webhook.TextMessageContent:
    bot.ReplyMessage(&messaging_api.ReplyMessageRequest{
        ReplyToken: e.ReplyToken,
        Messages: []messaging_api.MessageInterface{
            messaging_api.TextMessage{Text: message.Text},
        },
    })
```

### Pattern: Command Handler
```go
case webhook.TextMessageContent:
    switch {
    case strings.HasPrefix(message.Text, "/help"):
        replyHelp(bot, e.ReplyToken)
    case strings.HasPrefix(message.Text, "/weather"):
        replyWeather(bot, e.ReplyToken)
    default:
        replyUnknown(bot, e.ReplyToken)
    }
```

### Pattern: Media Download
```go
case webhook.ImageMessageContent:
    content, resp, err := blob.GetMessageContentWithHttpInfo(message.Id)
    if err != nil {
        log.Printf("Error: %v", err)
        return
    }
    defer resp.Body.Close()
    
    // Process image content
    saveImage(content, message.Id)
```

### Pattern: User Interaction
```go
// Send buttons
messaging_api.TemplateMessage{
    AltText: "Please select",
    Template: &messaging_api.ButtonsTemplate{
        Text: "What would you like to do?",
        Actions: []messaging_api.ActionInterface{
            &messaging_api.MessageAction{
                Label: "Option 1",
                Text:  "option1",
            },
            &messaging_api.MessageAction{
                Label: "Option 2",
                Text:  "option2",
            },
        },
    },
}

// Handle response
case webhook.MessageEvent:
    if message.Text == "option1" {
        // Handle option 1
    }
```

## Environment Variables Reference

| Variable | Required | Description | Example |
|----------|----------|-------------|---------|
| `LINE_CHANNEL_SECRET` | ✅ | Channel secret from LINE Developers | `abc123...` |
| `LINE_CHANNEL_TOKEN` | ✅ | Channel access token | `xyz789...` |
| `PORT` | ❌ | Server port (default: 5000) | `8080` |
| `APP_BASE_URL` | ⚠️ | Public HTTPS URL (Kitchen Sink only) | `https://example.com` |

⚠️ = Required for Kitchen Sink example

## Deployment Options

### Option 1: Heroku
```bash
# Create Procfile
echo "web: ./main" > Procfile

# Deploy
heroku create
heroku config:set LINE_CHANNEL_SECRET=your_secret
heroku config:set LINE_CHANNEL_TOKEN=your_token
git push heroku main
```

### Option 2: Google Cloud Run
```bash
# Create Dockerfile
FROM golang:1.24 AS builder
WORKDIR /app
COPY . .
RUN go build -o bot

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/bot /bot
CMD ["/bot"]

# Deploy
gcloud run deploy linebot --source .
```

### Option 3: Railway
```bash
railway init
railway add
railway up
```

## Troubleshooting

### Webhook Not Working
1. Check webhook URL is HTTPS
2. Verify webhook URL in LINE Developers Console
3. Check logs for signature validation errors
4. Ensure server is publicly accessible

### Messages Not Sending
1. Verify `LINE_CHANNEL_TOKEN` is correct
2. Check API response for errors
3. Use `*WithHttpInfo` methods to see detailed errors
4. Check LINE API status page

### 500 Errors
1. Check server logs
2. Verify all environment variables are set
3. Test signature validation
4. Check request payload format

## Additional Resources

- **LINE Developers**: https://developers.line.biz/
- **SDK Repository**: https://github.com/line/line-bot-sdk-go
- **API Reference**: https://developers.line.biz/en/reference/messaging-api/
- **Flex Message Simulator**: https://developers.line.biz/flex-simulator/
- **Icon Assets**: https://developers.line.biz/en/docs/messaging-api/icon-nickname-switch/

## Support

- **LINE Developers Community**: https://www.line-community.me/
- **Stack Overflow**: Tag `line-messaging-api`
- **GitHub Issues**: https://github.com/line/line-bot-sdk-go/issues
