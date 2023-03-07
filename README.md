# Twitch Chat Relay

This is a backend for my twitch chat relay. This operates [with the frontend](https://github.com/ShadiestGoat/twitch-chat-frontend)

## Configuration

This can be configured by editing the `.env` file. Checkout `template.env` for docs and format.

## Usage

Configure it, then build it using `go build`, and run it by doing `./twitch-chat-backend`. You will need the `PORT` for the frontend!

## Features and limits

This is only for the API, representation on the frontend may not be the same!

| supported              | not supported          |
|------------------------|------------------------|
| 7tv emotes             | non-7tv emotes         |
| bits                   | (TODO: add stuff here) |
| ACTION (/me)           | multiple channels      |
| twitch emotes          |                        |
| twitch emote blacklist |                        |
| twitch                 |                        |
| pronounDB pronouns     |                        |
| colors that contrast   |                        |

