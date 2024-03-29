module github.com/unit-01

go 1.18

// require github.com/bwmarrin/discordgo v0.25.0
replace github.com/bwmarrin/discordgo => /Users/quinton/Documents/Projects/Go/discordgo

require github.com/bwmarrin/discordgo v0.0.0-00010101000000-000000000000

require (
	github.com/sashabaranov/go-openai v1.4.2 // indirect
)

require (
	github.com/go-sql-driver/mysql v1.7.0
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/sashabaranov/go-gpt3 v0.0.0-20220803054136-8b463ceb2b74
	golang.org/x/crypto v0.0.0-20220722155217-630584e8d5aa // indirect
	golang.org/x/sys v0.0.0-20220731174439-a90be440212d // indirect
)
