package model

import "encoding/xml"

type Memory struct {
	Unit  *string `xml:"unit,attr"`
	Value *int    `xml:",chardata"`
}
type VCPU struct {
	Placement *string `xml:"placement,attr"`
	Value     *int    `xml:",chardata"`
}

type CPU struct {
	Mode     *string   `xml:"mode,attr"`
	Check    *string   `xml:"check,attr"`
	Topology Topology `xml:"topology"`
}

type Topology struct {
	Socket *int `xml:"sockets,attr"`
	Cores  *int `xml:"cores,attr"`
	Thread *int `xml:"threads,attr"`
}

type Resource struct {
	Partition *struct {
		Value *string `xml:",chardata"`
	} `xml:"partition"`
}

type OS struct {
	Boot *struct {
		Dev *string `xml:"dev,attr"`
	} `xml:"boot"`
	Type struct {
		Arch    *string `xml:"arch,attr"`
		Machine *string `xml:"machine,attr"`
		Value   *string `xml:",chardata"`
	} `xml:"type"`
}

type Features struct {
	Acpi   *struct{} `xml:"acpi"`
	Apic   *struct{} `xml:"apic"`
	Vmport *struct {
		State *string `xml:"state,attr"`
	} `xml:"vmport"`
}

type Disk struct {
	Driver *struct {
		Name *string `xml:"name,attr"`
		Type *string `xml:"type,attr"`
	} `xml:"driver"`
	Source *struct {
		File  *string  `xml:"file,attr"`
		Index *int `xml:"index,attr"`
	} `xml:"source"`
	Target *struct {
		Dev *string `xml:"dev,attr"`
		Bus *string `xml:"bus,attr"`
	} `xml:"target"`
	Address *struct {
		Type       *string `xml:"type,attr"`
		Controller *int    `xml:"controller,attr"`
		Bus        *int    `xml:"bus,attr"`
		Target     *int    `xml:"target,attr"`
		Unit       *int    `xml:"unit,attr"`
	} `xml:"address"`

	Type   *string `xml:"type,attr"`
	Device *string `xml:"device,attr"`
}

type Controller struct {
	Type  *string  `xml:"type,attr"`
	Index *int  `xml:"index,attr"`
	Model *string `xml:"model,attr"`

	Master *struct {
		Startport *int `xml:"startport,attr"`
	} `xml:"master"`
	Address *struct {
		Type     *string `xml:"type,attr"`
		Domain   *string `xml:"domain,attr"`
		Bus      *string `xml:"bus,attr"`
		Slot     *string `xml:"slot,attr"`
		Function *string `xml:"function,attr"`
		Multifunction *string `xml:"multifunction,attr"`
	} `xml:"address"`
}

type Interface struct {
	Type string `xml:"type,attr"`
	Mac *struct {
		Address *string `xml:"address,attr"`
	} `xml:"mac"`
	Source *struct {
		Dev  *string `xml:"dev,attr"`
		Mode *string `xml:"mode,attr"`
	} `xml:"source"`
	Model *struct {
		Type *string `xml:"type,attr"`
	} `xml:"model"`
	Address *struct {
		Type     *string `xml:"type,attr"`
		Domain   *string `xml:"domain,attr"`
		Bus      *string `xml:"bus,attr"`
		Slot     *string `xml:"slot,attr"`
		Function *string `xml:"function,attr"`
	} `xml:"address"`
}

type HostDev struct {
	Mode          *string `xml:"mode,attr"`
	Type          *string `xml:"type,attr"`
	Managed       *string `xml:"managed,attr"`
	SourceAddress *struct {
		Domain   *string `xml:"domain,attr"`
		Bus      *string `xml:"bus,attr"`
		Slot     *string `xml:"slot,attr"`
		Function *string `xml:"function,attr"`
	} `xml:"source>address"`
	Address *struct {
		Type     *string `xml:"type,attr"`
		Domain   *string `xml:"domain,attr"`
		Bus      *string `xml:"bus,attr"`
		Slot     *string `xml:"slot,attr"`
		Function *string `xml:"function,attr"`
	} `xml:"address"`
}

type Emulator struct {
	Value *string `xml:",chardata"`
}

type Sound struct {
	Model        *string `xml:"model,attr"`
	SoundAddress *struct {
		Type     *string `xml:"type,attr"`
		Domain   *string `xml:"domain,attr"`
		Bus      *string `xml:"bus,attr"`
		Slot     *string `xml:"slot,attr"`
		Function *string `xml:"function,attr"`
	} `xml:"address"`
}

type Video struct {
	Model  *struct {
		Ram 		int `xml:"ram,attr"`
		Vram 		int `xml:"vram,attr"`
		Vgamem 		int `xml:"vgamem,attr"`
		Heads 		int `xml:"heads,attr"`
		Type 		string `xml:"type,attr"`
		Primary 	string `xml:"primary,attr"`
	} `xml:"model"`
	Address  *struct {
		Type 		string `xml:"type,attr"`
		Domain 		string `xml:"domain,attr"`
		Bus 		string `xml:"bus,attr"`
		Slot 		string `xml:"slot,attr"`
		Function 	string `xml:"function,attr"`
	} `xml:"address"`
}

type Graphic struct {
	Type 			string `xml:"type,attr"`
	Autoport 		string `xml:"autoport,attr"`
	Listen *struct{
		Type 		string `xml:"type,attr"`
	} `xml:"listen"`
	Image *struct{
		Compression string `xml:"compression,attr"`
	} `xml:"image"`
}

type Channel struct {
	Type 			string `xml:"type,attr"`
	Target struct {
		Type 		string `xml:"type,attr"`
		Name 		string `xml:"name,attr"`
	} `xml:"target"`
	Address struct {
		Type 		string `xml:"type,attr"`
		Controller 	string `xml:"controller,attr"`
		Bus 		string `xml:"bus,attr"`
		Port 		string `xml:"port,attr"`
	} `xml:"address"`
}

type Memballoon struct {
	Model string `xml:"model,attr"`
	Address *struct {
		Type  		string `xml:"type,attr"`
		Domain  	string `xml:"domain,attr"`
		Bus  		string `xml:"bus,attr"`
		Slot  		string `xml:"slot,attr"`
		Function  	string `xml:"function,attr"`
	} `xml:"address"`
}
type PM struct {
	SuspendToMem *struct {
		Enable string `xml:"enabled,attr"`
	} `xml:"suspend-to-mem"`
	SuspendToDisk *struct {
		Enable string `xml:"enabled,attr"`
	} `xml:"suspend-to-disk"`
}
type OnPoweroff struct {
	Value     *string `xml:",chardata"`
}
type OnReboot struct {
	Value     *string `xml:",chardata"`
}
type OnCrash struct {
	Value     *string `xml:",chardata"`
}
type Input struct {
	Type string `xml:"type,attr"`
	Bus  string `xml:"bus,attr"`
}

type Clock struct {
	Offset 				string `xml:"offset,attr"`
	Timers []struct {
    	Name 			*string `xml:"name,attr"`
		Tickpolicy 		*string `xml:"tickpolicy,attr"`
    	Present 		*string `xml:"present,attr"`
	} `xml:"timer"`
}

type Domain struct {
	XMLName xml.Name `xml:"domain" yaml:"domain,inline"`
	Type    *string   `xml:"type,attr"`

	Name          *string  `xml:"name"`
	Uuid          *string  `xml:"uuid"`

	Memory        Memory   `xml:"memory"`
	CurrentMemory Memory   `xml:"currentMemory"`
	VCpu          VCPU     `xml:"vcpu"`
	OS            OS       `xml:"os"`
	Features      Features `xml:"features"`
	Cpu           CPU      `xml:"cpu"`
	Clock		  Clock    `xml:"clock"`

	OnReboot      *OnReboot   `xml:"on_reboot"`
	OnPoweroff    *OnPoweroff `xml:"on_poweroff"`
	OnCrash       *OnCrash    `xml:"on_crash"`
	PM 			  *PM		  `xml:"pm"`


	Emulator    Emulator     `xml:"devices>emulator"`
	Disk        []Disk       `xml:"devices>disk"`
	Controllers []Controller `xml:"devices>controller"`
	Interfaces  []Interface  `xml:"devices>interface"`
	Channel     *Channel		 `xml:"devices>channel"`
	Input	    []Input      `xml:"devices>input"`
	Graphic     *Graphic      `xml:"devices>graphics"`
	Sound       *Sound        `xml:"devices>sound"`
	Video       *Video		 `xml:"devices>video"`
	Hostdevs    []HostDev    `xml:"devices>hostdev"`
	Memballoon  *Memballoon   `xml:"devices>memballoon"`
}