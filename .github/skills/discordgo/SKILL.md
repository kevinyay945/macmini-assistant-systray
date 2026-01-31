---
name: discordgo
description: Expert guidance for building Discord bots with Go using the bwmarrin/discordgo library, including bot initialization, event handlers, slash commands, message components (buttons, select menus, modals), interactions, intents, voice handling, and webhooks. Use when building Discord bots, implementing Discord bot features, handling Discord events, creating slash commands, working with Discord API, or any Discord bot development tasks in Go.
license: MIT
---

# DiscordGo

Expert guidance for building Discord bots with the `github.com/bwmarrin/discordgo` library in Go.

## Installation

```bash
go get github.com/bwmarrin/discordgo
```

## Core Concepts

### Bot Session Initialization

Create a Discord session with bot token:

```go
import "github.com/bwmarrin/discordgo"

// Initialize with bot token (prefix with "Bot ")
dg, err := discordgo.New("Bot " + botToken)
if err != nil {
    log.Fatal("error creating Discord session:", err)
}

// Set intents for what events to receive
dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

// Open websocket connection
err = dg.Open()
if err != nil {
    log.Fatal("error opening connection:", err)
}
defer dg.Close()
```

### Gateway Intents

Specify what events your bot needs access to. Common intents:

- `discordgo.IntentsGuildMessages` - Guild message events
- `discordgo.IntentsDirectMessages` - DM events
- `discordgo.IntentsMessageContent` - Message content (privileged)
- `discordgo.IntentsGuilds` - Guild events
- `discordgo.IntentsGuildMembers` - Member events (privileged)
- `discordgo.IntentsGuildPresences` - Presence updates (privileged)

Combine with bitwise OR: `dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent`

**Note:** Privileged intents require approval in Discord Developer Portal.

### Event Handlers

Register handlers for Discord events using `AddHandler`:

```go
// Register before calling dg.Open()
dg.AddHandler(messageCreate)
dg.AddHandler(ready)

func ready(s *discordgo.Session, event *discordgo.Ready) {
    log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    // Ignore messages from the bot itself
    if m.Author.ID == s.State.User.ID {
        return
    }
    
    if m.Content == "ping" {
        s.ChannelMessageSend(m.ChannelID, "Pong!")
    }
}
```

### Graceful Shutdown

Keep bot running until interrupted:

```go
import (
    "os"
    "os/signal"
    "syscall"
)

// Wait for interrupt signal
sc := make(chan os.Signal, 1)
signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
<-sc

// Cleanup
dg.Close()
```

## Slash Commands

### Registering Slash Commands

```go
commands := []*discordgo.ApplicationCommand{
    {
        Name:        "hello",
        Description: "Say hello",
    },
    {
        Name:        "options",
        Description: "Command with options",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:        discordgo.ApplicationCommandOptionString,
                Name:        "text",
                Description: "Text to echo",
                Required:    true,
            },
            {
                Type:        discordgo.ApplicationCommandOptionInteger,
                Name:        "count",
                Description: "Number of times to repeat",
                Required:    false,
                MinValue:    &minValue, // float64(1.0)
                MaxValue:    10.0,
            },
        },
    },
}

// Register commands (guild-specific or globally)
registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
for i, cmd := range commands {
    // For specific guild: ApplicationCommandCreate(appID, guildID, cmd)
    // For global: ApplicationCommandCreate(appID, "", cmd)
    registered, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
    if err != nil {
        log.Fatalf("Cannot create '%v' command: %v", cmd.Name, err)
    }
    registeredCommands[i] = registered
}
```

### Handling Slash Command Interactions

```go
commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "hello": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Hello!",
            },
        })
    },
    "options": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        // Access options
        options := i.ApplicationCommandData().Options
        optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
        for _, opt := range options {
            optionMap[opt.Name] = opt
        }
        
        text := optionMap["text"].StringValue()
        count := 1
        if opt, ok := optionMap["count"]; ok {
            count = int(opt.IntValue())
        }
        
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Repeating '%s' %d times", text, count),
            },
        })
    },
}

// Register interaction handler
dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
        h(s, i)
    }
})
```

### Command Cleanup

Remove commands when shutting down:

```go
for _, cmd := range registeredCommands {
    err := s.ApplicationCommandDelete(s.State.User.ID, guildID, cmd.ID)
    if err != nil {
        log.Printf("Cannot delete '%v' command: %v", cmd.Name, err)
    }
}
```

## Message Components

### Buttons

```go
s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,
    Data: &discordgo.InteractionResponseData{
        Content: "Choose an option:",
        Components: []discordgo.MessageComponent{
            discordgo.ActionsRow{
                Components: []discordgo.MessageComponent{
                    discordgo.Button{
                        Label:    "Yes",
                        Style:    discordgo.SuccessButton,
                        CustomID: "btn_yes",
                    },
                    discordgo.Button{
                        Label:    "No",
                        Style:    discordgo.DangerButton,
                        CustomID: "btn_no",
                    },
                    discordgo.Button{
                        Label: "Documentation",
                        Style: discordgo.LinkButton,
                        URL:   "https://discord.com/developers/docs",
                        Emoji: &discordgo.ComponentEmoji{Name: "📜"},
                    },
                },
            },
        },
    },
})
```

**Button Styles:** `PrimaryButton`, `SecondaryButton`, `SuccessButton`, `DangerButton`, `LinkButton`

### Select Menus

```go
discordgo.ActionsRow{
    Components: []discordgo.MessageComponent{
        discordgo.SelectMenu{
            MenuType:    discordgo.StringSelectMenu,
            CustomID:    "select_choice",
            Placeholder: "Choose an option...",
            MinValues:   &minOne,  // *int
            MaxValues:   3,
            Options: []discordgo.SelectMenuOption{
                {
                    Label:       "Option 1",
                    Value:       "opt1",
                    Description: "First option",
                    Emoji: &discordgo.ComponentEmoji{Name: "1️⃣"},
                    Default:     false,
                },
                {
                    Label: "Option 2",
                    Value: "opt2",
                },
            },
        },
    },
}
```

**Select Menu Types:** `StringSelectMenu`, `UserSelectMenu`, `RoleSelectMenu`, `MentionableSelectMenu`, `ChannelSelectMenu`

### Modals

```go
s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseModal,
    Data: &discordgo.InteractionResponseData{
        CustomID: "modal_survey",
        Title:    "User Survey",
        Components: []discordgo.MessageComponent{
            discordgo.ActionsRow{
                Components: []discordgo.MessageComponent{
                    discordgo.TextInput{
                        CustomID:    "feedback",
                        Label:       "Your feedback",
                        Style:       discordgo.TextInputParagraph,
                        Placeholder: "Tell us what you think...",
                        Required:    true,
                        MaxLength:   1000,
                        MinLength:   10,
                    },
                },
            },
        },
    },
})
```

**Text Input Styles:** `TextInputShort` (single line), `TextInputParagraph` (multi-line)

### Handling Component Interactions

```go
componentHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "btn_yes": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "You clicked Yes!",
                Flags:   discordgo.MessageFlagsEphemeral, // Only visible to user
            },
        })
    },
    "select_choice": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        data := i.MessageComponentData()
        selected := data.Values // []string of selected values
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("You selected: %v", selected),
            },
        })
    },
    "modal_survey": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        data := i.ModalSubmitData()
        feedback := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("Thanks for feedback: %s", feedback),
            },
        })
    },
}

dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
    switch i.Type {
    case discordgo.InteractionApplicationCommand:
        if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
            h(s, i)
        }
    case discordgo.InteractionMessageComponent:
        if h, ok := componentHandlers[i.MessageComponentData().CustomID]; ok {
            h(s, i)
        }
    case discordgo.InteractionModalSubmit:
        if h, ok := componentHandlers[i.ModalSubmitData().CustomID]; ok {
            h(s, i)
        }
    }
})
```

## Interaction Response Types

- `InteractionResponsePong` - ACK a ping
- `InteractionResponseChannelMessageWithSource` - Respond with a message
- `InteractionResponseDeferredChannelMessageWithSource` - Defer response, edit later
- `InteractionResponseDeferredMessageUpdate` - Defer update to message
- `InteractionResponseUpdateMessage` - Update the message
- `InteractionResponseModal` - Respond with a modal

## Common Patterns

### Sending Messages

```go
// Simple message
s.ChannelMessageSend(channelID, "Hello!")

// Complex message with embeds
s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
    Content: "Check this out:",
    Embeds: []*discordgo.MessageEmbed{
        {
            Title:       "Embed Title",
            Description: "Description text",
            Color:       0x00ff00,
            Fields: []*discordgo.MessageEmbedField{
                {
                    Name:   "Field 1",
                    Value:  "Value 1",
                    Inline: true,
                },
            },
        },
    },
})
```

### Editing Messages

```go
// Edit interaction response
s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
    Content: &newContent,
})

// Edit regular message
s.ChannelMessageEdit(channelID, messageID, "Updated content")
```

### Deleting Messages

```go
s.ChannelMessageDelete(channelID, messageID)
```

### Getting Guild/Channel Info

```go
guild, err := s.Guild(guildID)
channel, err := s.Channel(channelID)
member, err := s.GuildMember(guildID, userID)
```

### Ephemeral Messages

Messages only visible to the interaction user:

```go
Data: &discordgo.InteractionResponseData{
    Content: "Only you can see this!",
    Flags:   discordgo.MessageFlagsEphemeral,
}
```

## Advanced Features

### Autocomplete

Handle autocomplete for slash command options:

```go
case discordgo.InteractionApplicationCommandAutocomplete:
    data := i.ApplicationCommandData()
    choices := []*discordgo.ApplicationCommandOptionChoice{
        {Name: "Apple", Value: "apple"},
        {Name: "Banana", Value: "banana"},
    }
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionApplicationCommandAutocompleteResult,
        Data: &discordgo.InteractionResponseData{
            Choices: choices,
        },
    })
```

### Context Menus

Create user/message context menu commands:

```go
{
    Name: "Get User Info",
    Type: discordgo.UserApplicationCommand,
}
{
    Name: "Quote Message",
    Type: discordgo.MessageApplicationCommand,
}
```

### Permissions

Set default permissions for commands:

```go
dmPermission := false
defaultPerms := int64(discordgo.PermissionManageServer)

{
    Name:                     "admin-command",
    Description:              "Admin only command",
    DefaultMemberPermissions: &defaultPerms,
    DMPermission:             &dmPermission,
}
```

### Localization

Add localized names/descriptions:

```go
{
    Name:        "hello",
    Description: "Say hello",
    NameLocalizations: &map[discordgo.Locale]string{
        discordgo.ChineseCN: "你好",
        discordgo.Spanish:   "hola",
    },
    DescriptionLocalizations: &map[discordgo.Locale]string{
        discordgo.ChineseCN: "打招呼",
        discordgo.Spanish:   "Decir hola",
    },
}
```

## Best Practices

1. **Always ignore bot's own messages** in message handlers to prevent loops
2. **Set appropriate intents** - only request what you need
3. **Clean up commands** on shutdown to avoid duplicates
4. **Use ephemeral messages** for error/info messages
5. **Defer long operations** - use `InteractionResponseDeferredChannelMessageWithSource` then edit
6. **Handle errors gracefully** in interaction handlers
7. **Use CustomID prefixes** to organize component handlers (e.g., `"btn_confirm_delete_123"`)

## Reference Materials

For detailed examples, see `references/examples.md` which contains complete code samples for:
- Basic message bot
- Slash commands with options
- Interactive components (buttons, selects, modals)
- Context menus
- Autocomplete
- Voice features
- Auto-moderation
