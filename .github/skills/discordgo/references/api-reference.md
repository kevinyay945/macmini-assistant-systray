# DiscordGo Quick Reference

Quick reference for common DiscordGo patterns and API methods.

## Session Methods

### Messages
- `s.ChannelMessageSend(channelID, content)` - Send simple message
- `s.ChannelMessageSendEmbed(channelID, embed)` - Send embed
- `s.ChannelMessageSendComplex(channelID, data)` - Send complex message
- `s.ChannelMessageEdit(channelID, messageID, content)` - Edit message
- `s.ChannelMessageDelete(channelID, messageID)` - Delete message
- `s.ChannelMessagesBulkDelete(channelID, messageIDs)` - Bulk delete

### Interactions
- `s.InteractionRespond(interaction, response)` - Respond to interaction
- `s.InteractionResponseEdit(interaction, data)` - Edit interaction response
- `s.InteractionResponseDelete(interaction)` - Delete interaction response

### Guilds & Channels
- `s.Guild(guildID)` - Get guild info
- `s.Channel(channelID)` - Get channel info
- `s.GuildChannels(guildID)` - List channels
- `s.GuildMember(guildID, userID)` - Get member info
- `s.GuildMembers(guildID, after, limit)` - List members

### Roles
- `s.GuildRoleCreate(guildID)` - Create role
- `s.GuildRoleEdit(guildID, roleID, name, color, hoist, permissions, mentionable)` - Edit role
- `s.GuildRoleDelete(guildID, roleID)` - Delete role
- `s.GuildMemberRoleAdd(guildID, userID, roleID)` - Add role to member
- `s.GuildMemberRoleRemove(guildID, userID, roleID)` - Remove role from member

### Application Commands
- `s.ApplicationCommandCreate(appID, guildID, command)` - Register command
- `s.ApplicationCommandEdit(appID, guildID, cmdID, command)` - Edit command
- `s.ApplicationCommandDelete(appID, guildID, cmdID)` - Delete command
- `s.ApplicationCommands(appID, guildID)` - List commands

### User & DMs
- `s.User(userID)` - Get user info
- `s.UserChannelCreate(userID)` - Create DM channel
- `s.UserUpdate(username, avatar)` - Update bot profile

## Event Types

Register handlers with `s.AddHandler(handlerFunc)`:

- `Ready` - Bot connected
- `MessageCreate` - New message
- `MessageUpdate` - Message edited
- `MessageDelete` - Message deleted
- `InteractionCreate` - Interaction received
- `GuildMemberAdd` - Member joined
- `GuildMemberRemove` - Member left
- `GuildMemberUpdate` - Member updated
- `PresenceUpdate` - Presence changed
- `VoiceStateUpdate` - Voice state changed

## Common Intents

```go
// Message content (privileged)
discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

// Guild events
discordgo.IntentsGuilds

// Member events (privileged)
discordgo.IntentsGuildMembers

// Presence (privileged)
discordgo.IntentsGuildPresences

// Reactions
discordgo.IntentsGuildMessageReactions

// Voice
discordgo.IntentsGuildVoiceStates

// DMs
discordgo.IntentsDirectMessages
```

## Interaction Types

```go
discordgo.InteractionPing                              // Ping
discordgo.InteractionApplicationCommand                // Slash command
discordgo.InteractionMessageComponent                  // Button/select
discordgo.InteractionApplicationCommandAutocomplete    // Autocomplete
discordgo.InteractionModalSubmit                       // Modal submission
```

## Response Types

```go
discordgo.InteractionResponsePong                              // ACK ping
discordgo.InteractionResponseChannelMessageWithSource          // Send message
discordgo.InteractionResponseDeferredChannelMessageWithSource  // Defer, edit later
discordgo.InteractionResponseDeferredMessageUpdate             // Defer update
discordgo.InteractionResponseUpdateMessage                     // Update message
discordgo.InteractionResponseModal                             // Show modal
discordgo.InteractionApplicationCommandAutocompleteResult      // Autocomplete results
```

## Message Flags

```go
discordgo.MessageFlagsEphemeral              // Only visible to user
discordgo.MessageFlagsSuppressEmbeds         // Don't show embeds
discordgo.MessageFlagsCrossposted            // Message crossposted
discordgo.MessageFlagsSourceMessageDeleted   // Source deleted
discordgo.MessageFlagsUrgent                 // System urgent
discordgo.MessageFlagsLoading                // Deferred response
```

## Button Styles

```go
discordgo.PrimaryButton    // Blurple
discordgo.SecondaryButton  // Grey
discordgo.SuccessButton    // Green
discordgo.DangerButton     // Red
discordgo.LinkButton       // Grey (requires URL)
```

## Component Types

```go
discordgo.ActionsRow         // Container for components
discordgo.Button             // Button
discordgo.SelectMenu         // Select menu
discordgo.TextInput          // Text input (modals only)
```

## Select Menu Types

```go
discordgo.StringSelectMenu       // String options
discordgo.UserSelectMenu         // User selection
discordgo.RoleSelectMenu         // Role selection
discordgo.MentionableSelectMenu  // User or role
discordgo.ChannelSelectMenu      // Channel selection
```

## Text Input Styles

```go
discordgo.TextInputShort      // Single line
discordgo.TextInputParagraph  // Multi-line
```

## Permission Constants

```go
discordgo.PermissionAdministrator
discordgo.PermissionManageGuild
discordgo.PermissionManageRoles
discordgo.PermissionManageChannels
discordgo.PermissionKickMembers
discordgo.PermissionBanMembers
discordgo.PermissionManageMessages
discordgo.PermissionSendMessages
discordgo.PermissionEmbedLinks
discordgo.PermissionAttachFiles
discordgo.PermissionMentionEveryone
discordgo.PermissionViewChannel
discordgo.PermissionConnect        // Voice
discordgo.PermissionSpeak          // Voice
discordgo.PermissionMuteMembers    // Voice
discordgo.PermissionDeafenMembers  // Voice
```

## Application Command Option Types

```go
discordgo.ApplicationCommandOptionSubCommand
discordgo.ApplicationCommandOptionSubCommandGroup
discordgo.ApplicationCommandOptionString
discordgo.ApplicationCommandOptionInteger
discordgo.ApplicationCommandOptionBoolean
discordgo.ApplicationCommandOptionUser
discordgo.ApplicationCommandOptionChannel
discordgo.ApplicationCommandOptionRole
discordgo.ApplicationCommandOptionMentionable
discordgo.ApplicationCommandOptionNumber
discordgo.ApplicationCommandOptionAttachment
```

## Embed Color Codes

```go
0x1ABC9C  // Turquoise
0x2ECC71  // Green
0x3498DB  // Blue
0x9B59B6  // Purple
0xE91E63  // Pink
0xF1C40F  // Yellow
0xE67E22  // Orange
0xE74C3C  // Red
0x95A5A6  // Light Grey
0x607D8B  // Dark Grey
0x11806A  // Dark Turquoise
0x1F8B4C  // Dark Green
0x206694  // Dark Blue
0x71368A  // Dark Purple
0xAD1457  // Dark Pink
0xC27C0E  // Dark Yellow
0xA84300  // Dark Orange
0x992D22  // Dark Red
```

## Common Patterns

### Parse Command Options to Map
```go
options := i.ApplicationCommandData().Options
optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
for _, opt := range options {
    optionMap[opt.Name] = opt
}
```

### Access Option Values
```go
text := optionMap["name"].StringValue()
num := optionMap["count"].IntValue()
user := optionMap["user"].UserValue(s)
channel := optionMap["channel"].ChannelValue(s)
role := optionMap["role"].RoleValue(s, guildID)
```

### Defer Long Operations
```go
s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
})

// Do long operation...

s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
    Content: &result,
})
```

### Get Message Author in Interaction
```go
// In guild
author := i.Member.User

// In DM
author := i.User
```

### Error Handling in Interactions
```go
if err != nil {
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseChannelMessageWithSource,
        Data: &discordgo.InteractionResponseData{
            Content: "An error occurred: " + err.Error(),
            Flags:   discordgo.MessageFlagsEphemeral,
        },
    })
    return
}
```
