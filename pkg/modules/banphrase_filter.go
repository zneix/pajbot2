package modules

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	twitch "github.com/gempir/go-twitch-irc"
	"github.com/pajlada/pajbot2/pkg"
	"github.com/pajlada/pajbot2/pkg/filters"
	"github.com/pajlada/pajbot2/pkg/utils"
)

type Pajbot1BanphraseFilter struct {
	server *server

	banphrases []pkg.Banphrase

	Sender pkg.Channel
}

func NewPajbot1BanphraseFilter(sender pkg.Channel) *Pajbot1BanphraseFilter {
	return &Pajbot1BanphraseFilter{
		server: &_server,
		Sender: sender,
	}
}

func (m *Pajbot1BanphraseFilter) loadPajbot1Banphrases() error {
	const queryF = `SELECT * FROM tb_banphrase`

	session := m.server.oldSession

	stmt, err := session.Prepare(queryF)
	if err != nil {
		return err
	}

	rows, err := stmt.Query()
	if err != nil {
		return err
	}

	for rows.Next() {
		var bp filters.Pajbot1Banphrase
		err = bp.LoadScan(rows)
		if err != nil {
			return err
		}

		if bp.Enabled {
			m.banphrases = append(m.banphrases, &bp)
		}
	}

	return nil
}

func (m *Pajbot1BanphraseFilter) Register() error {
	err := m.loadPajbot1Banphrases()
	if err != nil {
		return err
	}

	return nil
}

func (m Pajbot1BanphraseFilter) Name() string {
	return "Pajbot1BanphraseFilter"
}

type TimeoutData struct {
	FullMessage string
	Banphrase   pkg.Banphrase
	Username    string
	Channel     string
	Timestamp   time.Time
}

func (m Pajbot1BanphraseFilter) OnMessage(channel string, user twitch.User, message twitch.Message) error {
	originalVariations, lowercaseVariations, err := utils.MakeVariations(message.Text, true)
	if err != nil {
		return err
	}

	for _, bp := range m.banphrases {
		var variations *[]string

		if !bp.IsCaseSensitive() {
			variations = &lowercaseVariations
		} else {
			variations = &originalVariations
		}

		for _, variation := range *variations {
			if bp.Triggers(variation) {
				// fmt.Printf("Banphrase triggered: %#v\n", bp)
				if bp.IsAdvanced() && channel == "forsen" {
					lol := TimeoutData{
						FullMessage: message.Text,
						Banphrase:   bp,
						Username:    user.Username,
						Channel:     channel,
						Timestamp:   time.Now().UTC(),
					}
					c := m.server.redis.Pool.Get()
					bytes, _ := json.Marshal(&lol)
					c.Do("LPUSH", "pajbot2:timeouts", bytes)
					c.Close()
				}

				if channel == "krakenbul" && user.UserType == "" {
					reason := fmt.Sprintf("Matched banphrase with name '%s' and id '%d'", bp.GetName(), bp.GetID())
					m.Sender.Timeout(channel, user, bp.GetLength(), reason)
					// m.Sender.Say(channel, user.Username+" matched banphrase with name "+bp.GetName())
					log.Println("Matched banphrase with name \"" + bp.GetName() + "\"")
					// banphrase triggered
				}
				return nil
			}

			if !bp.IsAdvanced() {
				break
			}
		}
	}

	return nil
}