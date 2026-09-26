package discord

import (
	"fmt"
	"os"
	"strings"

	"service/database"
	"service/log"
	"service/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/samber/mo"
)

var session *discordgo.Session

const (
	WebName = "Mod Developer Branding"
)

const (
	colorPrimary   = 11241556
	colorSecondary = 11762602
	colorTertiary  = 12368721
)

type webhookCredentials struct {
	session *discordgo.Session
	id      string
	token   string
}

func getSession(private bool) mo.Result[webhookCredentials] {
	if session != nil {
		var id string
		var token string

		if private {
			id = os.Getenv("DISCORD_WH_ID_STAFF")
			if id == "" {
				return mo.Err[webhookCredentials](fmt.Errorf("discord staff webhook id variable is not defined!"))
			}

			token = os.Getenv("DISCORD_WH_TOKEN_STAFF")
			if token == "" {
				return mo.Err[webhookCredentials](fmt.Errorf("discord staff webhook token variable is not defined!"))
			}
		} else {
			id = os.Getenv("DISCORD_WH_ID")
			if id == "" {
				return mo.Err[webhookCredentials](fmt.Errorf("discord webhook id variable is not defined!"))
			}

			token = os.Getenv("DISCORD_WH_TOKEN")
			if token == "" {
				return mo.Err[webhookCredentials](fmt.Errorf("discord webhook token variable is not defined!"))
			}
		}

		return mo.Ok(webhookCredentials{session: session, id: id, token: token})
	} else {
		return mo.Err[webhookCredentials](fmt.Errorf("no discord session found"))
	}
}

func getDevHyperlink(dev string) string {
	return fmt.Sprintf("**[@%s](https://geode-sdk.org/mods?per_page=20&developer=%s&sort=recently_updated)**", dev, strings.ToLower(dev))
}

func WebhookAccept(img *utils.Img, staff *utils.User) mo.Result[bool] {
	credentialsRes := getSession(false)
	if credentialsRes.IsError() {
		return mo.Err[bool](credentialsRes.Error())
	}
	credentials := credentialsRes.MustGet()

	userRes := database.GetUser(img.UserID)
	if userRes.IsError() {
		return mo.Err[bool](userRes.Error())
	}
	u := userRes.MustGet()

	var mod string
	if staff != nil {
		mod = fmt.Sprintf("[@%s](https://www.github.com/%s/)", staff.Login, staff.Login)
	} else {
		mod = "<:ico:1325250328005967932> Developer is verified"
	}

	go func() {
		_, err := credentials.session.WebhookExecute(credentials.id, credentials.token, false, &discordgo.WebhookParams{
			Username: WebName,
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

	return mo.Ok(true)
}

func WebhookStaffSubmit(img *utils.Img) mo.Result[bool] {
	credentialsRes := getSession(true)
	if credentialsRes.IsError() {
		return mo.Err[bool](credentialsRes.Error())
	}
	credentials := credentialsRes.MustGet()

	userRes := database.GetUser(img.UserID)
	if userRes.IsError() {
		return mo.Err[bool](userRes.Error())
	}
	u := userRes.MustGet()

	go func() {
		_, err := credentials.session.WebhookExecute(credentials.id, credentials.token, false, &discordgo.WebhookParams{
			Username: WebName,
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

	return mo.Ok(true)
}

func init() {
	s, err := discordgo.New("")
	if err != nil {
		log.Error(err.Error())
		return
	}

	session = s
}
