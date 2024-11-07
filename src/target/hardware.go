package target

import (
	wmi "github.com/yusufpapurcu/wmi"
	"sryxen/utils"
)

func getCPU() (cpu utils.CPU, err error) {
	var cpus []utils.CPU
	err = wmi.Query("SELECT Name FROM Win32_Processor", &cpus)
	if err != nil {
		return
	}
	return cpus[0], nil
}

func getGPU() (gpus []utils.GPU, err error) {
	err = wmi.Query("SELECT Name, VideoProcessor, AdapterRAM FROM Win32_VideoController", &gpus)
	return
}

func getDisks() (disks []utils.Disk, err error) {
	err = wmi.Query("SELECT Name, FileSystem, FreeSpace, Size, VolumeName FROM Win32_LogicalDisk WHERE DriveType = 3", &disks)
	return
}

func getMotherboard() (mb utils.Motherboard, err error) {
	var mbs []utils.Motherboard
	err = wmi.Query("SELECT Manufacturer, Product, SerialNumber FROM Win32_BaseBoard WHERE Status = 'OK'", &mbs)
	if err != nil {
		return
	}
	return mbs[0], nil
}


func getUUID() (uuid string, err error) {
	var UUID []utils.UUID
	err = wmi.Query("SELECT UUID FROM Win32_ComputerSystemProduct", &UUID)
	if err != nil {
		return
	}
	return UUID[0].UUID, nil
}

func getMacAddress() (MAC string, err error) {
	var adapters []utils.NetAdapter
	err = wmi.Query("SELECT MACAddress FROM Win32_NetworkAdapter WHERE NetEnabled = True", &adapters)
	if err != nil {
		return
	}
	return adapters[0].MACAddress, nil
}

func GetComputerSpecifications() (pc utils.PC, err error) {
	cpu, err := getCPU()
	if err != nil {
		return
	}
	gpus, err := getGPU()
	if err != nil {
		return
	}
	disks, err := getDisks()
	if err != nil {
		return
	}
	motherboard, err := getMotherboard()
	if err != nil {
		return
	}
	uuid, err := getUUID()
	if err != nil {
		return
	}
	mac, err := getMacAddress()
	if err != nil {
		return
	}
	pc.CPU = cpu.Name
	pc.GPU = gpus
	pc.Motherboard = motherboard
	pc.Disks = disks
	pc.UUID = uuid
	pc.MacAddress = mac
	return
}
