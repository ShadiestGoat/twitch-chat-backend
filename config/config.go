package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/ShadiestGoat/twitch-chat-backend/log"
	"github.com/joho/godotenv"
)

type confItem struct {
	Env string

	// Has to be *string, *int, *float64, *bool, *[]string. Slices are space separated!
	// Bool is case insensitive and supports the following keys:
	// TRUE -> true,
	Res any

	// If not nil, is used to set the value of Res (Res must be a pointer!)
	Parser func(inp string) any

	Default  any
	Required bool

	// 'ITEM' is not set, so CONSEQUENCE
	Consequence string
}

var (
	TWITCH_CHANNEL_NAME = ""
	TWITCH_CHANNEL_ID   = ""
	PORT                = ""
	EMOTE_BLACKLIST     = []string{}
)

func Init() {
	godotenv.Load(".env")

	debugMention := ""
	debugWebhook := ""

	var confMap = []*confItem{
		{
			Env:      "PORT",
			Res:      &PORT,
			Required: true,
		},
		{
			Env:      "TWITCH_CHANNEL_ID",
			Res:      &TWITCH_CHANNEL_ID,
			Required: true,
		},
		{
			Env:      "TWITCH_CHANNEL_NAME",
			Res:      &TWITCH_CHANNEL_NAME,
			Required: true,
		},
		{
			Env: "DEBUG_MENTION",
			Res: &debugMention,
		},
		{
			Env:         "DEBUG_WEBHOOK",
			Res:         &debugWebhook,
			Consequence: "it will not send debug messages to discord",
		},
		{
			Env:         "EMOTE_BLACKLIST",
			Res:         &EMOTE_BLACKLIST,
			Consequence: "there shall not be any blacklisted emotes",
		},
	}

	for _, opt := range confMap {
		handleKey(opt)
	}

	log.Init(debugMention, debugWebhook)
}

func handleKey(opt *confItem) {
	item := os.Getenv(opt.Env)

	if item == "" {
		if opt.Required {
			log.Fatal("'%s' is a needed variable, but is not present! Please read the README.md file for more info.", opt.Env)
		}

		if opt.Default == nil {
			if opt.Consequence != "" {
				log.Warn("'%s' is not set, so %s!", opt.Env, opt.Consequence)
			}
			return
		}

		switch opt.Res.(type) {
		case *[]string:
			*(opt.Res.(*[]string)) = opt.Default.([]string)
		case *string:
			*(opt.Res.(*string)) = opt.Default.(string)
		case *bool:
			*(opt.Res.(*bool)) = opt.Default.(bool)
		case *float64:
			*(opt.Res.(*float64)) = opt.Default.(float64)
		case *int:
			*(opt.Res.(*int)) = opt.Default.(int)
		default:
			log.Fatal("Couldn't set the default value for '%s'", opt.Env)
		}

		return
	}

	switch resP := opt.Res.(type) {
	case *[]string:
		*resP = strings.Split(item, " ")
	case *string:
		*resP = item
	case *bool:
		v, err := strconv.ParseBool(item)
		if err != nil {
			log.Fatal("'%s' is not a valid bool value!", opt.Env)
		}

		*resP = v
	case *float64:
		v, err := strconv.ParseFloat(item, 64)
		if err != nil {
			log.Fatal("'%s' is not a valid float value!", opt.Env)
		}

		*resP = v
	case *int:
		v, err := strconv.Atoi(item)
		if err != nil {
			log.Fatal("'%s' is not a valid int value!", opt.Env)
		}

		*resP = v
	default:
		if opt.Parser == nil {
			log.Fatal("Unknown type for '%s'", opt.Env)
		}

		v := opt.Parser(item)
		*(resP.(*any)) = v
	}
}
