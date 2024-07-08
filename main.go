package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	websiteURL = "https://ziu.gov.pl/api/ZIUZW"
)

var cache = expirable.NewLRU[string, string](5, nil, time.Minute*25)
var cacheResult, _ = lru.New[string, bool](128)

type Result struct {
	ID    int    `json:"ID"`
	Nazwa string `json:"Nazwa"`
	Sesje []struct {
		ID          int    `json:"ID"`
		Nazwa       string `json:"Nazwa"`
		ObcaSesjaID int    `json:"ObcaSesjaID"`
		RokSzkolny  string `json:"RokSzkolny"`
		TypSesji    struct {
			ID    int    `json:"ID"`
			Nazwa string `json:"Nazwa"`
		} `json:"TypSesji"`
		DataOtwarciaSesji     int `json:"DataOtwarciaSesji"`
		DataZamknieciaSesji   int `json:"DataZamknieciaSesji"`
		DataPublikacjiWynikow int `json:"DataPublikacjiWynikow"`
		Wyniki                []struct {
			ID         int `json:"ID"`
			Rspo       int `json:"Rspo"`
			OkeID      int `json:"OkeID"`
			Uzytkownik struct {
				ID               int    `json:"ID"`
				Login            string `json:"Login"`
				CzyIstniejeLogin bool   `json:"CzyIstniejeLogin"`
			} `json:"Uzytkownik"`
			EgzaminWSesji struct {
				ID        int `json:"ID"`
				EgzaminID int `json:"EgzaminID"`
				Egzamin   struct {
					ID            int    `json:"ID"`
					Nazwa         string `json:"Nazwa"`
					TypDeklaracji struct {
						ID              int    `json:"ID"`
						Nazwa           string `json:"Nazwa"`
						TypZgloszeniaID int    `json:"TypZgloszeniaID"`
						TypZgloszenia   struct {
							ID    int    `json:"ID"`
							Nazwa string `json:"Nazwa"`
							Kod   string `json:"Kod"`
						} `json:"TypZgloszenia"`
					} `json:"TypDeklaracji"`
					FormaZdawania struct {
						ID    int    `json:"ID"`
						Nazwa string `json:"Nazwa"`
						Kod   string `json:"Kod"`
					} `json:"FormaZdawania"`
					PoziomZdawania struct {
						ID    int    `json:"ID"`
						Nazwa string `json:"Nazwa"`
					} `json:"PoziomZdawania"`
					Kod           string `json:"Kod"`
					JezykZdawania struct {
						ID                int    `json:"ID"`
						Nazwa             string `json:"Nazwa"`
						Kod               string `json:"Kod"`
						BlokEgzaminacyjny struct {
							ID    int    `json:"ID"`
							Nazwa string `json:"Nazwa"`
						} `json:"BlokEgzaminacyjny"`
					} `json:"JezykZdawania"`
					KodKrem           string `json:"KodKrem"`
					BlokEgzaminacyjny struct {
						ID    int    `json:"ID"`
						Nazwa string `json:"Nazwa"`
					} `json:"BlokEgzaminacyjny"`
					CzyZwolnienieZEgzaminu bool `json:"CzyZwolnienieZEgzaminu"`
					CzyWycofany            bool `json:"CzyWycofany"`
					NazwaSystemuID         int  `json:"NazwaSystemuID"`
					ObcyEgzaminID          int  `json:"ObcyEgzaminID"`
					HashID                 int  `json:"HashID"`
				} `json:"Egzamin"`
				SesjaID int `json:"SesjaID"`
				Sesja   struct {
					ID             int    `json:"ID"`
					Nazwa          string `json:"Nazwa"`
					ObcaSesjaID    int    `json:"ObcaSesjaID"`
					NazwaSystemuID int    `json:"NazwaSystemuID"`
					RokSzkolny     string `json:"RokSzkolny"`
					TypSesji       struct {
						ID    int    `json:"ID"`
						Nazwa string `json:"Nazwa"`
					} `json:"TypSesji"`
					DataOtwarciaSesji                 int `json:"DataOtwarciaSesji"`
					DataZamknieciaSesji               int `json:"DataZamknieciaSesji"`
					DataPublikacjiWynikow             int `json:"DataPublikacjiWynikow"`
					DataZakonczeniaWprowadzaniaDanych int `json:"DataZakonczeniaWprowadzaniaDanych"`
					TypZgloszenia                     struct {
						ID    int    `json:"ID"`
						Nazwa string `json:"Nazwa"`
						Kod   string `json:"Kod"`
					} `json:"TypZgloszenia"`
					SesjaGlownaID int  `json:"SesjaGlownaID"`
					CzyCovid      bool `json:"CzyCovid"`
					Aktualna      bool `json:"Aktualna"`
				} `json:"Sesja"`
				Termin int `json:"Termin"`
			} `json:"EgzaminWSesji"`
			KodArkusza             string  `json:"KodArkusza"`
			MiejsceWydania         string  `json:"MiejsceWydania"`
			DataWydania            int     `json:"DataWydania"`
			NumerWydanegoDokumentu string  `json:"NumerWydanegoDokumentu"`
			Centyle                int     `json:"Centyle"`
			Procent                int     `json:"Procent"`
			MaxPunkty              float64 `json:"MaxPunkty"`
			UzyskanePunkty         float64 `json:"UzyskanePunkty"`
			StanWyniku             struct {
				ID    int    `json:"ID"`
				Nazwa string `json:"Nazwa"`
			} `json:"StanWyniku"`
			StanZdawaniaEgzaminu struct {
				ID    int    `json:"ID"`
				Nazwa string `json:"Nazwa"`
			} `json:"StanZdawaniaEgzaminu"`
			CzyObowiazkowy bool `json:"CzyObowiazkowy"`
		} `json:"Wyniki"`
		PlacowkaNazwa string `json:"PlacowkaNazwa"`
	} `json:"Sesje"`
}

func main() {
	password := os.Getenv("ZIU_PASSWORD")
	if password == "" {
		panic("ZIU_PASSWORD environment variable not set")
	}
	username := os.Getenv("ZIU_USERNAME")
	if username == "" {
		panic("ZIU_USERNAME environment variable not set")
	}
	discordWebhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if discordWebhookURL == "" {
		panic("DISCORD_WEBHOOK_URL environment variable not set")
	}

	_, err := refreshToken(username, password)
	if err != nil {
		fmt.Println("Error logging in:", err)
		return
	}

	for {
		token, ok := cache.Get("token")
		if !ok {
			token, err = refreshToken(username, password)
			if err != nil {
				fmt.Println("Error logging in:", err)
				time.Sleep(time.Second * 20)
				continue
			}
		}

		fmt.Println("Fetching result...")
		jsonResponse, err := getResult(token)
		if err != nil {
			fmt.Println("Error fetching result:", err)
			time.Sleep(time.Second * 20)
			continue
		}
		var result []Result
		err = json.Unmarshal([]byte(jsonResponse), &result)
		if err != nil {
			fmt.Println("Error unmarshalling result:", err)
			time.Sleep(time.Second * 20)
			continue
		}
		message := formatResult(result)

		fmt.Println("Sending to Discord...")
		err = sendToDiscord(discordWebhookURL, message, jsonResponse)
		if err != nil {
			fmt.Println("Error sending to Discord:", err)
		}
		fmt.Println("Sent to Discord, waiting for next check...")
		time.Sleep(time.Minute)
	}
}

func refreshToken(username, password string) (string, error) {
	response, err := http.Post(websiteURL+"/uzytkownik/login", "application/json",
		strings.NewReader(fmt.Sprintf(`{"Login":"%s","Zeton":"%s"}`, username, password)))
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status: %s", response.Status)
	}

	token := response.Header.Get("X-Token")
	cache.Add("token", token)
	return token, nil
}

func getResult(token string) (string, error) {
	req, err := http.NewRequest("GET", websiteURL+"/Wynik", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-Token", token)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected response status: %s", response.Status)
	}

	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func formatResult(result []Result) string {
	message := ""
	for _, r := range result {
		for _, s := range r.Sesje {
			for i, w := range s.Wyniki {
				ok := cacheResult.Contains(fmt.Sprintf("result_%d_%d", s.ID, w.ID))
				if ok {
					fmt.Printf("Skipping result %d from session %d\n", w.ID, s.ID)
					continue
				}
				if i == 0 {
					message += fmt.Sprintf("# %s\n", s.Nazwa)
					message += fmt.Sprintf("Placówka: %s\n", s.PlacowkaNazwa)
				}
				message += fmt.Sprintf("## Egzamin: %s (poz. %s) (%s)\n", w.EgzaminWSesji.Egzamin.Nazwa,
					w.EgzaminWSesji.Egzamin.PoziomZdawania.Nazwa,
					w.EgzaminWSesji.Egzamin.FormaZdawania.Nazwa)
				message += fmt.Sprintf("Data wydania dokumentu: %s\n", time.Unix(int64(w.DataWydania), 0).Format("2006-01-02 15:04:05"))
				message += fmt.Sprintf("Numer wydanego dokumentu: **%s**\n", w.NumerWydanegoDokumentu)
				message += fmt.Sprintf("Data egzaminu: %s\n", time.Unix(int64(w.EgzaminWSesji.Termin), 0).Format("2006-01-02 15:04:05"))
				message += fmt.Sprintf("Kod arkusza: %s\n", w.KodArkusza)
				message += fmt.Sprintf("Centyle: %d\n", w.Centyle)
				message += fmt.Sprintf("\n**Procent: %d**\n\n", w.Procent)
				message += fmt.Sprintf("Punkty: %.2f/%.2f\n", w.UzyskanePunkty, w.MaxPunkty)
				message += "\n"

				cacheResult.Add(fmt.Sprintf("result_%d_%d", s.ID, w.ID), true)
			}
		}
	}
	return message
}

func sendToDiscord(webhookURL, message string, jsonData string) (err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if message == "" {
		fmt.Println("No message to send")
		return nil
	}

	filesIndex := 0
	if jsonData != "" {
		part, err := writer.CreateFormFile(fmt.Sprintf("files[%d]", filesIndex), "message.json")
		if err != nil {
			return err
		}
		_, err = part.Write([]byte(jsonData))
		if err != nil {
			return err
		}
		filesIndex++
	}
	if len(message) > 2000 {
		part, err := writer.CreateFormFile(fmt.Sprintf("files[%d]", filesIndex), "message.txt")
		if err != nil {
			return err
		}
		_, err = part.Write([]byte(jsonData))
		if err != nil {
			return err
		}
		filesIndex++
	} else {
		err = writer.WriteField("content", message)
		if err != nil {
			return err
		}
	}

	err = writer.Close()
	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewReader(body.Bytes()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Printf("Discord response body: %s\n", string(bodyBytes))
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}

	return nil
}
