# DiscordGo Complete Examples

Comprehensive code examples for common Discord bot patterns.

## Table of Contents

- [Basic Message Bot (Ping-Pong)](#basic-message-bot-ping-pong)
- [Slash Commands with Options](#slash-commands-with-options)
- [Interactive Components](#interactive-components)
- [Modals and Forms](#modals-and-forms)
- [Context Menus](#context-menus)
- [Autocomplete](#autocomplete)
- [Message Embeds](#message-embeds)
- [DM Handling](#dm-handling)

## Basic Message Bot (Ping-Pong)

Simple bot that responds to "ping" with "Pong!":

```go
package main

import (
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/bwmarrin/discordgo"
)

var Token string

func init() {
    flag.StringVar(&Token, "t", "", "Bot Token")
    flag.Parse()
}

func main() {
    dg, err := discordgo.New("Bot " + Token)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return
    }
    
    dg.AddHandler(messageCreate)
    dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent
    
    err = dg.Open()
    if err != nil {
        fmt.Println("error opening connection,", err)
        return
    }
    
    fmt.Println("Bot is now running. Press CTRL-C to exit.")
    sc := make(chan os.Signal, 1)
    signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-sc
    
    dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == s.State.User.ID {
        return
    }
    
    if m.Content == "ping" {
        s.ChannelMessageSend(m.ChannelID, "Pong!")
    }
    
    if m.Content == "pong" {
        s.ChannelMessageSend(m.ChannelID, "Ping!")
    }
}
```

## Slash Commands with Options

Full example with command registration, handlers, and cleanup:

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/bwmarrin/discordgo"
)

var (
    GuildID        = flag.String("guild", "", "Test guild ID")
    BotToken       = flag.String("token", "", "Bot access token")
    RemoveCommands = flag.Bool("rmcmd", true, "Remove commands after shutdown")
)

var s *discordgo.Session

func init() {
    flag.Parse()
    
    var err error
    s, err = discordgo.New("Bot " + *BotToken)
    if err != nil {
        log.Fatalf("Invalid bot parameters: %v", err)
    }
}

var (
    minValue = 1.0
    commands = []*discordgo.ApplicationCommand{
        {
            Name:        "greet",
            Description: "Greet a user",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionUser,
                    Name:        "user",
                    Description: "User to greet",
                    Required:    true,
                },
                {
                    Type:        discordgo.ApplicationCommandOptionString,
                    Name:        "message",
                    Description: "Custom greeting message",
                    Required:    false,
                },
            },
        },
        {
            Name:        "echo",
            Description: "Echo a message",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionString,
                    Name:        "text",
                    Description: "Text to echo",
                    Required:    true,
                },
                {
                    Type:        discordgo.ApplicationCommandOptionInteger,
                    Name:        "times",
                    Description: "Number of times to repeat",
                    MinValue:    &minValue,
                    MaxValue:    5,
                    Required:    false,
                },
            },
        },
    }
    
    commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
        "greet": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            options := i.ApplicationCommandData().Options
            optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
            for _, opt := range options {
                optionMap[opt.Name] = opt
            }
            
            user := optionMap["user"].UserValue(s)
            message := "Hello"
            if msg, ok := optionMap["message"]; ok {
                message = msg.StringValue()
            }
            
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: fmt.Sprintf("%s, %s!", message, user.Mention()),
                },
            })
        },
        "echo": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
            options := i.ApplicationCommandData().Options
            optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
            for _, opt := range options {
                optionMap[opt.Name] = opt
            }
            
            text := optionMap["text"].StringValue()
            times := 1
            if t, ok := optionMap["times"]; ok {
                times = int(t.IntValue())
            }
            
            response := ""
            for i := 0; i < times; i++ {
                response += text + "\n"
            }
            
            s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
                Type: discordgo.InteractionResponseChannelMessageWithSource,
                Data: &discordgo.InteractionResponseData{
                    Content: response,
                },
            })
        },
    }
)

func main() {
    s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
            h(s, i)
        }
    })
    
    s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
        log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
    })
    
    err := s.Open()
    if err != nil {
        log.Fatalf("Cannot open the session: %v", err)
    }
    
    log.Println("Registering commands...")
    registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
    for i, v := range commands {
        cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *GuildID, v)
        if err != nil {
            log.Panicf("Cannot create '%v' command: %v", v.Name, err)
        }
        registeredCommands[i] = cmd
    }
    
    defer s.Close()
    
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    log.Println("Press Ctrl+C to exit")
    <-stop
    
    if *RemoveCommands {
        log.Println("Removing commands...")
        for _, v := range registeredCommands {
            err := s.ApplicationCommandDelete(s.State.User.ID, *GuildID, v.ID)
            if err != nil {
                log.Panicf("Cannot delete '%v' command: %v", v.Name, err)
            }
        }
    }
    
    log.Println("Gracefully shutting down.")
}
```

## Interactive Components

Buttons and select menus:

```go
var componentsHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "show_buttons": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Do you understand buttons?",
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
                                Label: "Learn More",
                                Style: discordgo.LinkButton,
                                URL:   "https://discord.com/developers/docs",
                                Emoji: &discordgo.ComponentEmoji{Name: "📚"},
                            },
                        },
                    },
                },
            },
        })
        if err != nil {
            log.Println("Error responding:", err)
        }
    },
    "btn_yes": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Great! You understand buttons! 🎉",
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
    },
    "btn_no": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "No worries! Check out the Discord docs to learn more.",
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
    },
    "show_select": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        minValues := 1
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: "Choose your favorite programming language:",
                Components: []discordgo.MessageComponent{
                    discordgo.ActionsRow{
                        Components: []discordgo.MessageComponent{
                            discordgo.SelectMenu{
                                MenuType:    discordgo.StringSelectMenu,
                                CustomID:    "select_lang",
                                Placeholder: "Select a language...",
                                MinValues:   &minValues,
                                MaxValues:   3,
                                Options: []discordgo.SelectMenuOption{
                                    {
                                        Label:       "Go",
                                        Value:       "go",
                                        Description: "Fast, compiled language",
                                        Emoji:       &discordgo.ComponentEmoji{Name: "🐹"},
                                        Default:     true,
                                    },
                                    {
                                        Label:       "Python",
                                        Value:       "python",
                                        Description: "Easy to learn",
                                        Emoji:       &discordgo.ComponentEmoji{Name: "🐍"},
                                    },
                                    {
                                        Label:       "JavaScript",
                                        Value:       "js",
                                        Description: "Web development",
                                        Emoji:       &discordgo.ComponentEmoji{Name: "💻"},
                                    },
                                    {
                                        Label: "Rust",
                                        Value: "rust",
                                        Emoji: &discordgo.ComponentEmoji{Name: "🦀"},
                                    },
                                },
                            },
                        },
                    },
                },
            },
        })
    },
    "select_lang": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        data := i.MessageComponentData()
        langs := data.Values
        
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("You selected: %v", langs),
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
    },
}
```

## Modals and Forms

Create interactive forms:

```go
var modalHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "feedback": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseModal,
            Data: &discordgo.InteractionResponseData{
                CustomID: "modal_feedback",
                Title:    "Feedback Form",
                Components: []discordgo.MessageComponent{
                    discordgo.ActionsRow{
                        Components: []discordgo.MessageComponent{
                            discordgo.TextInput{
                                CustomID:    "subject",
                                Label:       "Subject",
                                Style:       discordgo.TextInputShort,
                                Placeholder: "Brief subject...",
                                Required:    true,
                                MaxLength:   100,
                            },
                        },
                    },
                    discordgo.ActionsRow{
                        Components: []discordgo.MessageComponent{
                            discordgo.TextInput{
                                CustomID:    "message",
                                Label:       "Message",
                                Style:       discordgo.TextInputParagraph,
                                Placeholder: "Your detailed feedback...",
                                Required:    true,
                                MaxLength:   2000,
                                MinLength:   10,
                            },
                        },
                    },
                },
            },
        })
        if err != nil {
            log.Println("Error showing modal:", err)
        }
    },
    "modal_feedback": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        data := i.ModalSubmitData()
        
        subject := ""
        message := ""
        
        for _, row := range data.Components {
            actionsRow := row.(*discordgo.ActionsRow)
            for _, component := range actionsRow.Components {
                input := component.(*discordgo.TextInput)
                if input.CustomID == "subject" {
                    subject = input.Value
                } else if input.CustomID == "message" {
                    message = input.Value
                }
            }
        }
        
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("**Feedback Received!**\nSubject: %s\nMessage: %s", subject, message),
                Flags:   discordgo.MessageFlagsEphemeral,
            },
        })
        
        // Optionally send to a logging channel
        // s.ChannelMessageSend(logChannelID, fmt.Sprintf("New feedback from %s:\n**%s**\n%s", 
        //     i.Member.User.Username, subject, message))
    },
}
```

## Context Menus

User and message context menus:

```go
var contextCommands = []*discordgo.ApplicationCommand{
    {
        Name: "User Info",
        Type: discordgo.UserApplicationCommand,
    },
    {
        Name: "Quote Message",
        Type: discordgo.MessageApplicationCommand,
    },
}

var contextHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
    "User Info": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        data := i.ApplicationCommandData()
        user := data.Resolved.Users[data.TargetID]
        
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Embeds: []*discordgo.MessageEmbed{
                    {
                        Title: "User Information",
                        Fields: []*discordgo.MessageEmbedField{
                            {Name: "Username", Value: user.Username, Inline: true},
                            {Name: "ID", Value: user.ID, Inline: true},
                            {Name: "Bot", Value: fmt.Sprint(user.Bot), Inline: true},
                        },
                        Thumbnail: &discordgo.MessageEmbedThumbnail{
                            URL: user.AvatarURL("256"),
                        },
                        Color: 0x00ff00,
                    },
                },
                Flags: discordgo.MessageFlagsEphemeral,
            },
        })
    },
    "Quote Message": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        data := i.ApplicationCommandData()
        msg := data.Resolved.Messages[data.TargetID]
        
        s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
            Type: discordgo.InteractionResponseChannelMessageWithSource,
            Data: &discordgo.InteractionResponseData{
                Content: fmt.Sprintf("> %s\n— %s", msg.Content, msg.Author.Mention()),
            },
        })
    },
}
```

## Autocomplete

Dynamic option suggestions:

```go
var autocompleteCommands = []*discordgo.ApplicationCommand{
    {
        Name:        "search",
        Description: "Search for something",
        Options: []*discordgo.ApplicationCommandOption{
            {
                Type:         discordgo.ApplicationCommandOptionString,
                Name:         "query",
                Description:  "Search query",
                Required:     true,
                Autocomplete: true,
            },
        },
    },
}

func handleAutocomplete(s *discordgo.Session, i *discordgo.InteractionCreate) {
    data := i.ApplicationCommandData()
    
    // Get the current user input
    var query string
    for _, opt := range data.Options {
        if opt.Name == "query" {
            query = opt.StringValue()
        }
    }
    
    // Generate suggestions based on input
    suggestions := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}
    choices := []*discordgo.ApplicationCommandOptionChoice{}
    
    for _, suggestion := range suggestions {
        if strings.HasPrefix(strings.ToLower(suggestion), strings.ToLower(query)) {
            choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
                Name:  suggestion,
                Value: suggestion,
            })
        }
        if len(choices) >= 25 { // Discord limit
            break
        }
    }
    
    s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionApplicationCommandAutocompleteResult,
        Data: &discordgo.InteractionResponseData{
            Choices: choices,
        },
    })
}

// In main handler:
func main() {
    s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
        switch i.Type {
        case discordgo.InteractionApplicationCommand:
            // Handle commands
        case discordgo.InteractionApplicationCommandAutocomplete:
            handleAutocomplete(s, i)
        }
    })
}
```

## Message Embeds

Rich formatted messages:

```go
func sendEmbed(s *discordgo.Session, channelID string) {
    embed := &discordgo.MessageEmbed{
        Title:       "Embed Title",
        Description: "This is an embedded message",
        URL:         "https://example.com",
        Color:       0x00ff00, // Green
        Timestamp:   time.Now().Format(time.RFC3339),
        Footer: &discordgo.MessageEmbedFooter{
            Text:    "Footer text",
            IconURL: "https://example.com/icon.png",
        },
        Thumbnail: &discordgo.MessageEmbedThumbnail{
            URL: "https://example.com/thumb.png",
        },
        Image: &discordgo.MessageEmbedImage{
            URL: "https://example.com/image.png",
        },
        Author: &discordgo.MessageEmbedAuthor{
            Name:    "Author Name",
            IconURL: "https://example.com/author.png",
        },
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:   "Field 1",
                Value:  "Value 1",
                Inline: true,
            },
            {
                Name:   "Field 2",
                Value:  "Value 2",
                Inline: true,
            },
            {
                Name:   "Non-inline Field",
                Value:  "This field spans the full width",
                Inline: false,
            },
        },
    }
    
    s.ChannelMessageSendEmbed(channelID, embed)
}
```

## DM Handling

Send and handle direct messages:

```go
// Send a DM
func sendDM(s *discordgo.Session, userID string, content string) error {
    channel, err := s.UserChannelCreate(userID)
    if err != nil {
        return err
    }
    
    _, err = s.ChannelMessageSend(channel.ID, content)
    return err
}

// Handle DM messages
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
    if m.Author.ID == s.State.User.ID {
        return
    }
    
    // Check if it's a DM
    channel, err := s.Channel(m.ChannelID)
    if err != nil {
        return
    }
    
    if channel.Type == discordgo.ChannelTypeDM {
        s.ChannelMessageSend(m.ChannelID, "Thanks for your DM! This is a direct message response.")
        return
    }
    
    // Handle guild messages
    if m.Content == "!dm" {
        err := sendDM(s, m.Author.ID, "This is a direct message!")
        if err != nil {
            s.ChannelMessageSend(m.ChannelID, "Failed to send DM. Make sure your DMs are open!")
        }
    }
}
```
