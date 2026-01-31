# DiscordGo Best Practices & Troubleshooting

## Best Practices

### 1. Bot Token Security

**Never** hardcode tokens in your source code:

```go
// ❌ Bad
token := "MTk5Mz..."

// ✅ Good - Use environment variables
token := os.Getenv("DISCORD_BOT_TOKEN")

// ✅ Good - Use command-line flags
var token string
flag.StringVar(&token, "token", "", "Bot token")
flag.Parse()
```

### 2. Ignore Bot's Own Messages

Always check if message author is the bot to prevent infinite loops:

```go
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Always do this first
    if m.Author.ID == s.State.User.ID {
        return
    }
    // ... rest of handler
}
```

### 3. Request Only Needed Intents

Only request intents your bot actually needs:

```go
// ❌ Bad - Too broad
dg.Identify.Intents = discordgo.IntentsAll

// ✅ Good - Specific intents
dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent
```

### 4. Use Ephemeral Messages for Errors

Error messages should typically be ephemeral (only visible to user):

```go
Data: &discordgo.InteractionResponseData{
    Content: "Error: Invalid input!",
    Flags:   discordgo.MessageFlagsEphemeral,
}
```

### 5. Defer Long Operations

Interactions must be responded to within 3 seconds. Defer if processing takes longer:

```go
// Defer immediately
s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
})

// Do long operation
result := processLongOperation()

// Edit deferred response
s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
    Content: &result,
})
```

### 6. Clean Up Commands

Remove commands on shutdown to prevent duplicates:

```go
defer func() {
    for _, cmd := range registeredCommands {
        s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
    }
}()
```

### 7. Use CustomID Prefixes

Organize component handlers with prefixed CustomIDs:

```go
// Use prefixes for different action types
CustomID: "btn_confirm_delete_" + itemID
CustomID: "select_category_" + contextID
CustomID: "modal_feedback_" + userID

// Parse in handler
if strings.HasPrefix(customID, "btn_confirm_delete_") {
    itemID := strings.TrimPrefix(customID, "btn_confirm_delete_")
    // Handle deletion
}
```

### 8. Validate User Input

Always validate data from command options and modals:

```go
text := optionMap["message"].StringValue()
if len(text) == 0 {
    // Handle error
}
if len(text) > 2000 {
    text = text[:2000]
}
```

### 9. Handle Permissions

Check permissions before performing actions:

```go
func hasPermission(s *discordgo.Session, guildID, userID string, permission int64) bool {
    member, err := s.GuildMember(guildID, userID)
    if err != nil {
        return false
    }
    
    for _, roleID := range member.Roles {
        role, err := s.State.Role(guildID, roleID)
        if err != nil {
            continue
        }
        if role.Permissions&permission != 0 {
            return true
        }
    }
    return false
}
```

### 10. Proper Error Logging

Log errors with context for debugging:

```go
if err != nil {
    log.Printf("Failed to send message to channel %s: %v", channelID, err)
    return
}
```

## Common Issues & Solutions

### Issue: "Missing Access" Error

**Problem:** Bot doesn't have required permissions.

**Solutions:**
1. Check bot role permissions in guild settings
2. Ensure bot role is above roles it needs to manage
3. Verify channel-specific permission overrides
4. Check if required intents are enabled in Developer Portal

### Issue: "Invalid Form Body" Error

**Problem:** Command/component structure is invalid.

**Solutions:**
1. Ensure all commands have name and description
2. Check that required fields are provided
3. Validate option types match their values
4. Ensure CustomIDs don't exceed 100 characters
5. Check that component limits aren't exceeded (5 rows, 5 buttons per row, 25 options per select)

### Issue: Message Content is Empty

**Problem:** `m.Content` is always empty despite receiving messages.

**Solutions:**
1. Enable `MESSAGE_CONTENT` privileged intent in Developer Portal
2. Add intent to code: `dg.Identify.Intents |= discordgo.IntentsMessageContent`
3. Note: Only required for bot messages, slash commands work without it

### Issue: Commands Not Appearing

**Problem:** Slash commands don't show up in Discord.

**Solutions:**
1. Wait up to 1 hour for global commands to propagate
2. Use guild-specific commands for testing (instant)
3. Check command name follows Discord rules (lowercase, no spaces)
4. Verify bot has `applications.commands` scope
5. Delete duplicate commands: check for multiple registrations

### Issue: Interaction Failed

**Problem:** "This interaction failed" message appears.

**Solutions:**
1. Respond within 3 seconds or defer the interaction
2. Check for panics in handler code
3. Ensure `InteractionRespond` is called exactly once
4. Verify response structure is valid
5. Check logs for error messages

### Issue: Bot Disconnects Frequently

**Problem:** Bot keeps disconnecting and reconnecting.

**Solutions:**
1. Check network connectivity
2. Verify token is valid
3. Ensure no duplicate sessions (one token = one session)
4. Check for rate limiting
5. Review Discord API status page

### Issue: Voice Connection Fails

**Problem:** Bot can't connect to voice channels.

**Solutions:**
1. Ensure bot has CONNECT and SPEAK permissions
2. Check voice channel user limit
3. Verify voice intents: `discordgo.IntentsGuildVoiceStates`
4. Review voice connection code for proper initialization

### Issue: Rate Limited

**Problem:** Getting 429 (Too Many Requests) errors.

**Solutions:**
1. DiscordGo has built-in rate limiting - respect it
2. Don't spam API calls in loops
3. Use bulk operations when possible (e.g., `BulkDelete`)
4. Cache data instead of repeated API calls
5. Review Discord rate limit documentation

### Issue: Can't DM User

**Problem:** Unable to send direct messages to users.

**Solutions:**
1. User may have DMs disabled from server members
2. Bot needs to share a server with the user
3. Check for error and handle gracefully:
```go
channel, err := s.UserChannelCreate(userID)
if err != nil {
    // Handle: user DMs closed
    return
}
```

## Debugging Tips

### Enable Debug Logging

```go
dg.LogLevel = discordgo.LogDebug
dg.Debug = true
```

### Print Interaction Data

```go
func debugInteraction(i *discordgo.InteractionCreate) {
    log.Printf("Type: %v", i.Type)
    log.Printf("Data: %+v", i.Data)
    if i.Member != nil {
        log.Printf("Member: %v", i.Member.User.Username)
    }
}
```

### Validate Component Structure

```go
// Check component counts
func validateComponents(components []discordgo.MessageComponent) error {
    if len(components) > 5 {
        return errors.New("too many action rows (max 5)")
    }
    for _, row := range components {
        if ar, ok := row.(discordgo.ActionsRow); ok {
            if len(ar.Components) > 5 {
                return errors.New("too many components in row (max 5)")
            }
        }
    }
    return nil
}
```

### Test Commands Locally

Use guild commands for instant updates during development:

```go
// Development
cmd, err := s.ApplicationCommandCreate(appID, testGuildID, command)

// Production
cmd, err := s.ApplicationCommandCreate(appID, "", command) // Global
```

## Performance Tips

### 1. Use State Cache

DiscordGo maintains a state cache. Use it instead of API calls:

```go
// ✅ Fast - uses cache
guild, err := s.State.Guild(guildID)

// ⚠️ Slower - API call
guild, err := s.Guild(guildID)
```

### 2. Batch Operations

Group related operations:

```go
// ❌ Multiple API calls
for _, msgID := range messageIDs {
    s.ChannelMessageDelete(channelID, msgID)
}

// ✅ Single API call
s.ChannelMessagesBulkDelete(channelID, messageIDs)
```

### 3. Reuse Sessions

Don't create multiple sessions for the same bot:

```go
// ❌ Bad - creates multiple connections
for i := 0; i < 10; i++ {
    dg, _ := discordgo.New("Bot " + token)
    dg.Open()
}

// ✅ Good - one session
dg, _ := discordgo.New("Bot " + token)
dg.Open()
defer dg.Close()
```

### 4. Optimize Event Handlers

Only register handlers you need:

```go
// ❌ Registers for all events
dg.AddHandler(func(s *discordgo.Session, e *discordgo.Event) {
    // Generic handler
})

// ✅ Specific handlers
dg.AddHandler(messageCreate)
dg.AddHandler(interactionCreate)
```

## Security Considerations

### 1. Validate Permissions

Always verify user has required permissions before executing commands:

```go
defaultPerms := int64(discordgo.PermissionManageServer)
command := &discordgo.ApplicationCommand{
    Name:                     "admin",
    Description:              "Admin command",
    DefaultMemberPermissions: &defaultPerms,
}
```

### 2. Sanitize User Input

Never trust user input directly:

```go
// Escape mentions to prevent abuse
content = strings.ReplaceAll(content, "@everyone", "@\u200beveryone")
content = strings.ReplaceAll(content, "@here", "@\u200bhere")
```

### 3. Rate Limit User Actions

Implement cooldowns for resource-intensive commands:

```go
var commandCooldowns = make(map[string]time.Time)

func checkCooldown(userID, command string, duration time.Duration) bool {
    key := userID + ":" + command
    if lastUse, exists := commandCooldowns[key]; exists {
        if time.Since(lastUse) < duration {
            return false // On cooldown
        }
    }
    commandCooldowns[key] = time.Now()
    return true // Can use
}
```

### 4. Validate File Uploads

Check file sizes and types when handling attachments:

```go
for _, attachment := range m.Attachments {
    if attachment.Size > 8*1024*1024 { // 8MB
        s.ChannelMessageSend(m.ChannelID, "File too large!")
        return
    }
}
```

## Testing Best Practices

### 1. Use Test Guild

Create a dedicated test server for development:

```go
var testGuildID = "your-test-guild-id"

if isDevelopment {
    s.ApplicationCommandCreate(appID, testGuildID, cmd)
} else {
    s.ApplicationCommandCreate(appID, "", cmd)
}
```

### 2. Mock Interactions

Create helper functions for testing:

```go
func createMockInteraction(commandName string) *discordgo.InteractionCreate {
    return &discordgo.InteractionCreate{
        Interaction: &discordgo.Interaction{
            Type: discordgo.InteractionApplicationCommand,
            Data: discordgo.ApplicationCommandInteractionData{
                Name: commandName,
            },
        },
    }
}
```

### 3. Test Error Paths

Ensure error handling works:

```go
// Test with invalid inputs
// Test with missing permissions
// Test with rate limiting
// Test with network failures
```
