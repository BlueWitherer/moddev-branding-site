package discord

import (
	"fmt"
	"os"
	"strings"

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

func getDevHyperlink(dev string) string {
	return fmt.Sprintf("**[@%s](https://geode-sdk.org/mods?per_page=20&developer=%s&sort=recently_updated)**", dev, strings.ToLower(dev))
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
		mod = "<:ico:1325250328005967932> Developer is verified"
	}

	go func() {
		_, err = s.WebhookExecute(id, token, false, &discordgo.WebhookParams{
			Username:  WebName,
			AvatarURL: WebAvatar,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "✅ New Developer Branding",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Developer",
							Value:  getDevHyperlink(u.Login),
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

		if err != nil {
			log.Error(err.Error())
		}
	}()

	return nil
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

	go func() {
		_, err = s.WebhookExecute(id, token, false, &discordgo.WebhookParams{
			Username:  WebName,
			AvatarURL: WebAvatar,
			Embeds: []*discordgo.MessageEmbed{
				{
					Title: "🕑 Branding Submission",
					Fields: []*discordgo.MessageEmbedField{
						{
							Name:   "Developer",
							Value:  getDevHyperlink(u.Login),
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

		if err != nil {
			log.Error(err.Error())
		}
	}()

	return nil
}

func init() {
	s, err := discordgo.New("")
	if err != nil {
		log.Error(err.Error())
		return
	}

	session = s
}
