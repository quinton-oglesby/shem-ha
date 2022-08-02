package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	gogpt3 "github.com/sashabaranov/go-gpt3"
)

// Creating a struct to hold the two tokens.
type Tokens struct {
	DiscordToken string
	GPT3Token    string
}

// Creating a struct that will hold all the GPT3 parameters.
type Parameters struct {
	Chance float64
	Length int64
}

// Creating a struct that will hold the channel array of allowed channels.
type Channels struct {
	Channels []string
}

// Globalizing the structs that hold this important data.
var tokens Tokens
var parameters Parameters
var channels Channels

// Global variable to hold the regex string, because why not?
var re *regexp.Regexp

// Main functions.
func main() {

	// Retrieve the tokens from the tokens.json file.
	tokensFile, err := os.ReadFile("tokens.json")
	if err != nil {
		log.Fatal("COULD NOT READ 'tokens.json' FILE: ", err)
	}

	// Unmarshal the tokens from tokensFile.
	json.Unmarshal(tokensFile, &tokens)

	// Retrieve the parameters from the GPT3Parameters.json file.
	parametersFile, err := os.ReadFile("parameters.json")
	if err != nil {
		log.Fatal("COULD NOT READ 'parameters.json' FILE: ", err)
	}

	// Unmarshal the tokens from the gp3ParametersFile.
	json.Unmarshal(parametersFile, &parameters)

	// Retrieve the channels from the channels.json file.
	channelsFile, err := os.ReadFile("channels.json")
	if err != nil {
		log.Fatal("COULD NOT READ 'channels.json' FILE: ", err)
	}

	// Unmarshal the channels from channelsFile.
	json.Unmarshal(channelsFile, &channels)

	// Compile regex string.
	re, err = regexp.Compile(`([\w+]+\:\/\/)?([\w\d-]+\.)*[\w-]+[\.\:]\w+([\/\?\=\&\#\.]?[\w-]+)*\/?`)
	if err != nil {
		log.Fatal("COULD NOT COMPILE REGEX: ", err)
	}

	// Create a new Discord session using the provided bot token.
	session, err := discordgo.New("Bot " + tokens.DiscordToken)
	if err != nil {
		log.Fatal("ERROR CREATING DISCORD SESSION:", err)
	}

	// Identify that we want all intents.
	session.Identify.Intents = discordgo.IntentsAll

	// Now we open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		log.Fatal("ERROR OPENING WEBSOCKET:", err)
	}

	// // Making a map of registered commands.
	// registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))

	// // Looping through the commands array and registering them.
	// // https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.ApplicationCommandCreate
	// for i, command := range commands {
	// 	registered_command, err := session.ApplicationCommandCreate(session.State.User.ID, "336297387863703552", command)
	// 	if err != nil {
	// 		log.Printf("CANNOT CREATE '%v' COMMAND: %v", command.Name, err)
	// 	}
	// 	registeredCommands[i] = registered_command
	// }

	// Looping through the array of interaction handlers and adding them to the session.
	session.AddHandler(func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		if handler, ok := commandHandlers[interaction.ApplicationCommandData().Name]; ok {
			handler(session, interaction)
		}
	})

	// Add the messageCreate handler to the session.
	session.AddHandler(messageCreate)

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// // Lopping through the registeredCommands array and deleting all the commands.
	// for _, v := range registeredCommands {
	// 	err := session.ApplicationCommandDelete(session.State.User.ID, "1001077854936760352", v.ID)
	// 	if err != nil {
	// 		log.Printf("CANNOT DELETE '%v' COMMAND: %v", v.Name, err)
	// 	}
	// }

	// Cleanly close down the Discord session.
	session.Close()
}

// Decalaring default member permission.
var defaultMemberPermissions int64 = discordgo.PermissionManageServer

// Declaring min and max values of the chance command option.
var minChanceValue float64 = 0
var maxChanceValue float64 = 100

// Declaring the max value allowed for a response.
var minLengthValue float64 = 60
var maxLengthValue float64 = 512

// Creating an array of commands to register.
//https://pkg.go.dev/github.com/bwmarrin/discordgo#ApplicationCommand
var commands = []*discordgo.ApplicationCommand{
	{
		Name:                     "test",
		Description:              "This is just a test command!",
		DefaultMemberPermissions: &defaultMemberPermissions,
	},
	{
		Name:                     "list_channels",
		Description:              "This command shows all the channels that Shem-Ha is allowed to post in.",
		DefaultMemberPermissions: &defaultMemberPermissions,
	},
	{
		Name:                     "echo",
		Description:              "This echoes your text to the specified channel as Shem-Ha.",
		DefaultMemberPermissions: &defaultMemberPermissions,

		// Registering the option available for this command.
		// https://pkg.go.dev/github.com/bwmarrin/discordgo#ApplicationCommandOption
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "This is the specified channel that you want to echo your message to.",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "text",
				Description: "This is the text that you want Shem-Ha to echo.",
				Required:    true,
			},
		},
	},
	{
		Name:                     "get_chance",
		Description:              "This returns the value of the chance that Shem-Ha will respond to a message.",
		DefaultMemberPermissions: &defaultMemberPermissions,
	},
	{
		Name:                     "get_length",
		Description:              "This returns the maximum length of a response from Shem-Ha in tokens. A token is about 4 characters.",
		DefaultMemberPermissions: &defaultMemberPermissions,
	},
	{
		Name:                     "set_chance",
		Description:              "This sets the value of the chance that Shem-Ha will respond to a message.",
		DefaultMemberPermissions: &defaultMemberPermissions,
		// Registering the option available for this command.
		// https://pkg.go.dev/github.com/bwmarrin/discordgo#ApplicationCommandOption
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionNumber,
				Name:        "percentage",
				Description: "This value is the chance that Shem-Ha will respond to a message, must be between 0 and 100.",
				Required:    true,
				MinValue:    &minChanceValue,
				MaxValue:    maxChanceValue,
			},
		},
	},
	{
		Name:                     "set_length",
		Description:              "This sets the maximum length of a response from Shem-Ha in tokens. A token is about 4 characters.",
		DefaultMemberPermissions: &defaultMemberPermissions,
		// Registering the option available for this command.
		// https://pkg.go.dev/github.com/bwmarrin/discordgo#ApplicationCommandOption
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionInteger,
				Name:        "tokens",
				Description: "This is the maximum response length in tokens. A token is about 4 characters.",
				Required:    true,
				MinValue:    &minLengthValue,
				MaxValue:    maxLengthValue,
			},
		},
	},
	{
		Name:                     "pop_channel",
		Description:              "This removes a channel from the list of channels Shem-Ha is allowed to reply in.",
		DefaultMemberPermissions: &defaultMemberPermissions,
		// Registering the option available for this command.
		// https://pkg.go.dev/github.com/bwmarrin/discordgo#ApplicationCommandOption
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel that you want to remove from the list of approved channels.",
				Required:    true,
			},
		},
	},
	{
		Name:                     "append_channel",
		Description:              "This adds a channel to the list of channels Shem-Ha is allowed to reply in.",
		DefaultMemberPermissions: &defaultMemberPermissions,
		// Registering the option available for this command.
		// https://pkg.go.dev/github.com/bwmarrin/discordgo#ApplicationCommandOption
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "channel",
				Description: "The channel that you want to add to the list of approved channels.",
				Required:    true,
			},
		},
	},
}

// Creating a map of event handlers to respond to application commands.
// https://pkg.go.dev/github.com/bwmarrin/discordgo#EventHandler
var commandHandlers = map[string]func(session *discordgo.Session, interaction *discordgo.InteractionCreate){
	"test": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Responding to the interaction.
		//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Congrats on using the test command!",
			},
		})
	},
	"echo": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Grabbing the channel ID and the content of the message to echo.
		channel := interaction.ApplicationCommandData().Options[0].ChannelValue(session)
		content := interaction.ApplicationCommandData().Options[1].StringValue()
		msg, err := session.ChannelMessageSend(channel.ID, content)
		if err != nil {
			log.Printf("COULD NOT SEND MESSAGE '%v': %v", msg, err)
		}

		// Responding to the interaction.
		//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully sent '%v' to channel '%v'", content, channel.Name),
			},
		})
	},
	"get_chance": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Responding to the interaction.
		//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("The current response chance is %v percent.", parameters.Chance),
			},
		})
	},
	"get_length": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Responding to the interaction.
		//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("The current response length is %v tokens.", parameters.Length),
			},
		})
	},
	"set_chance": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		parameters.Chance = interaction.ApplicationCommandData().Options[0].FloatValue()

		// Marshall the new parameters to save.
		jsonBytes, err := json.Marshal(parameters)
		if err != nil {
			log.Println("ERROR MARSHALING JSON: ", err)

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("FAILED TO UPDATE CHANCE: %v", err),
				},
			})

			return
		}

		// Save updated parameters into parameters.json.
		err = os.WriteFile("parameters.json", jsonBytes, 0644)
		if err != nil {
			log.Panicln("ERROR SAVING JSON: ", err)

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("FAILED TO UPDATE CHANCE: %v", err),
				},
			})

			return
		}

		// Responding to the interaction.
		//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully updated the response chance. The reponse chance is now %v percent.", parameters.Chance),
			},
		})
	},
	"set_length": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		parameters.Length = interaction.ApplicationCommandData().Options[0].IntValue()

		// Marshall the new parameters to save.
		jsonBytes, err := json.Marshal(parameters)
		if err != nil {
			log.Println("ERROR MARSHALING JSON: ", err)

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("FAILED TO UPDATE LENGTH: %v", err),
				},
			})

			return
		}

		// Save updated parameters into parameters.json.
		err = os.WriteFile("parameters.json", jsonBytes, 0644)
		if err != nil {
			log.Println("ERROR SAVING JSON: ", err)

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("FAILED TO UPDATE LENGTH: %v", err),
				},
			})

			return
		}

		// Responding to the interaction.
		//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully updated the response length. The reponse length is now %v tokens.", parameters.Length),
			},
		})
	},
	"pop_channel": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {
		// Snagging the target channel ID.
		targetChannelName := interaction.ApplicationCommandData().Options[0].ChannelValue(session).Name
		targetChannelID := interaction.ApplicationCommandData().Options[0].ChannelValue(session).ID

		// Checking if channel is already in the list of approved channels.
		if !stringInArray(targetChannelID, channels.Channels) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Channel '%v' is not in the list of approved channels.", targetChannelName),
				},
			})

			return
		}

		// Remove channel from the list of channels allowed.
		channels.Channels = removeStringFromArray(targetChannelID, channels.Channels)

		// Marshall the new channels to save.
		jsonBytes, err := json.Marshal(channels)
		if err != nil {
			log.Println("ERROR MARSHALING JSON: ", err)

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("FAILED TO UPDATE CHANNELS: %v", err),
				},
			})

			return
		}

		// Save updated parameters into parameters.json.
		err = os.WriteFile("channels.json", jsonBytes, 0644)
		if err != nil {
			log.Println("ERROR SAVING JSON: ", err)

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("FAILED TO UPDATE CHANNELS: %v", err),
				},
			})

			return
		}

		// Responding to the interaction.
		//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully removed '%v' from the list of channels I am allowed to respond in.", targetChannelName),
			},
		})
	},
	"append_channel": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

		// Snagging the target channel ID.
		targetChannelName := interaction.ApplicationCommandData().Options[0].ChannelValue(session).Name
		targetChannelID := interaction.ApplicationCommandData().Options[0].ChannelValue(session).ID

		// Checking if channel is already in the list of approved channels.
		if stringInArray(targetChannelID, channels.Channels) {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Channel '%v' is already in the list of approved channels.", targetChannelName),
				},
			})

			return
		}

		// Add channel to the list of channels allowed.
		channels.Channels = append(channels.Channels, targetChannelID)

		// Marshall the new channels to save.
		jsonBytes, err := json.Marshal(channels)
		if err != nil {
			log.Println("ERROR MARSHALING JSON: ", err)

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("FAILED TO UPDATE CHANNELS: %v", err),
				},
			})

			return
		}

		// Save updated parameters into parameters.json.
		err = os.WriteFile("channels.json", jsonBytes, 0644)
		if err != nil {
			log.Println("ERROR SAVING JSON: ", err)

			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("FAILED TO UPDATE CHANNELS: %v", err),
				},
			})

			return
		}

		// Responding to the interaction.
		//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
		session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Successfully added '%v' to the list of channels I am allowed to respond in.", targetChannelName),
			},
		})
	},
	"list_channels": func(session *discordgo.Session, interaction *discordgo.InteractionCreate) {

		if len(channels.Channels) > 0 {
			chnls := ""
			for _, channelID := range channels.Channels {
				channel, err := session.Channel(channelID)
				if err != nil {
					log.Println("ERROR RETREIVING CHANNELS: ", err)

					session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
						Type: discordgo.InteractionResponseChannelMessageWithSource,
						Data: &discordgo.InteractionResponseData{
							Content: fmt.Sprintf("FAILED TO GET CHANNELS: %v", err),
						},
					})
					return
				}

				chnls += channel.Name + "\n"
			}

			// Responding to the interaction.
			//https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.InteractionRespond
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: chnls,
				},
			})
		} else {
			session.InteractionRespond(interaction.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "CHANNEL LIST IS EMPTY",
				},
			})
		}

	},
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	// Ignore all messages that were created by the bot itself.
	if message.Author.ID == session.State.User.ID {
		return
	}

	// Ignore all messages with the discriminator #0000 (Webhooks).
	if message.WebhookID != "" {
		return
	}

	// Filter out all URLs in the message.
	message_content := re.ReplaceAllString(message.Content, "")

	// Ultimately ignore all messages with no content in them.
	if message_content == "" {
		return
	}

	// Check if the bot is allowed to respond in this channel.
	contains := stringInArray(message.ChannelID, channels.Channels)
	if contains {
		startChatLog := fmt.Sprintf(`The following is a conversation with an AI assistant named Shem-Ha. Shem-Ha acts like an arrogant goddess.
		%v: Hello. My name is %v.
		Shem-Ha: I am Shem-Ha. What do you want human?
		%v: %v
		Shem-Ha: `,
			message.Author.Username, message.Author.Username, message.Author.Username, message_content)

		// Craeting and seeding the random number generator.
		random := rand.New(rand.NewSource(time.Now().UnixNano()))

		// Generating the chance to repond to the message.
		chance := random.Float64()

		// Logging the chance to repond to the message.
		log.Println("CHANCE: ", chance*100.0)
		if chance*100 < parameters.Chance*1.0 {
			// Creating the GPT3 client.
			client := gogpt3.NewClient(tokens.GPT3Token)
			ctx := context.Background()

			// Building a completion request from GPT3.
			stops := []string{"\n", message.Author.Username}
			req := gogpt3.CompletionRequest{
				MaxTokens:        int(parameters.Length),
				Prompt:           startChatLog,
				Stop:             stops,
				Temperature:      1.0,
				TopP:             1.0,
				FrequencyPenalty: 1,
				PresencePenalty:  1,
				BestOf:           1,
			}

			response, err := client.CreateCompletion(ctx, "davinci", req)
			if err != nil {
				log.Println("COULD NOT COMPLETE A GPT3 COMPLETION: ", err)
				return
			}
			res := response.Choices[0].Text

			// https://pkg.go.dev/github.com/bwmarrin/discordgo#Session.ChannelMessageSendComplex
			_, err = session.ChannelMessageSendComplex(message.ChannelID, &discordgo.MessageSend{
				Content:   res,
				Reference: message.Reference(),
				AllowedMentions: &discordgo.MessageAllowedMentions{
					Parse: nil,
				},
			})
			if err != nil {
				log.Printf("COULD NOT REPLY TO %v: %v", message, err)
			}
		}
	}
}

// Function to check if a string is in an array, returns true or false.
func stringInArray(str string, list []string) bool {
	for _, i := range list {
		if i == str {
			return true
		}
	}

	return false
}

// Function to remove a string from an array, returning the newly updated array.
func removeStringFromArray(str string, list []string) []string {
	for i, j := range list {
		if j == str {
			return append(list[:i], list[i+1:]...)
		}
	}

	return list
}
