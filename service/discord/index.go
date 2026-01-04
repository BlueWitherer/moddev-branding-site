package discord

import (
	"fmt"
	"os"

	"service/database"
	"service/log"
	"service/utils"

	"github.com/bwmarrin/discordgo"
)

var session *discordgo.Session

const (
	WebName   = "Mod Developer Branding"
	WebAvatar = "https://github.com/BlueWitherer/ModDevBranding/blob/master/logo.png?raw=true"
)

const (
	colorPrimary   = 11241556
	colorSecondary = 11762602
	colorTertiary  = 12368721
)

func getSession(private bool) (*discordgo.Session, string, string, error) {
	if session != nil {
		var id string
		var token string

		if private {
			id = os.Getenv("DISCORD_WH_ID_STAFF")
			if id == "" {
				return nil, "", "", fmt.Errorf("discord staff webhook id variable is not defined!")
			}

			token = os.Getenv("DISCORD_WH_TOKEN_STAFF")
			if token == "" {
				return nil, "", "", fmt.Errorf("discord staff webhook token variable is not defined!")
			}
		} else {
			id = os.Getenv("DISCORD_WH_ID")
			if id == "" {
				return nil, "", "", fmt.Errorf("discord webhook id variable is not defined!")
			}

			token = os.Getenv("DISCORD_WH_TOKEN")
			if token == "" {
				return nil, "", "", fmt.Errorf("discord webhook token variable is not defined!")
			}
		}

		return session, id, token, nil
	} else {
		return nil, "", "", fmt.Errorf("no discord session found")
	}
}

func WebhookAccept(img *utils.Img, staff *utils.User) error {
	s, id, token, err := getSession(false)
	if err != nil {
		return err
	}

	u, err := database.GetUser(img.UserID)
	if err != nil {
		return err
	}

	var mod string
	if staff != nil {
		mod = fmt.Sprintf("[@%s](https://www.github.com/%s/)", staff.Login, staff.Login)
	} else {
		mod = "Developer is verified"
	}

	_, err = s.WebhookExecute(id, token, true, &discordgo.WebhookParams{
		Username:  WebName,
		AvatarURL: WebAvatar,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "✅ New Developer Branding",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Developer",
						Value:  fmt.Sprintf("**[@%s](https://www.github.com/%s/)**", u.Login, u.Login),
						Inline: true,
					},
					{
						Name:   "Moderator",
						Value:  mod,
						Inline: true,
					},
				},
				Color: colorPrimary,
				Image: &discordgo.MessageEmbedImage{
					URL:      img.ImageURL,
					ProxyURL: img.ImageURL,
				},
			},
		},
	})

	return err
}

func WebhookStaffSubmit(img *utils.Img) error {
	s, id, token, err := getSession(true)
	if err != nil {
		return err
	}

	u, err := database.GetUser(img.UserID)
	if err != nil {
		return err
	}

	_, err = s.WebhookExecute(id, token, true, &discordgo.WebhookParams{
		Username:  WebName,
		AvatarURL: WebAvatar,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title: "🕑 Branding Submission",
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:   "Developer",
						Value:  fmt.Sprintf("**[@%s](https://www.github.com/%s/)**", u.Login, u.Login),
						Inline: true,
					},
				},
				Color: colorTertiary,
				Image: &discordgo.MessageEmbedImage{
					URL:      img.ImageURL,
					ProxyURL: img.ImageURL,
				},
			},
		},
	})

	return err
}

func init() {
	s, err := discordgo.New("")
	if err != nil {
		log.Error(err.Error())
		return
	}

	session = s
}
