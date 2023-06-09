package virtdaemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"test/internal/libvirt"
	qemuhypervisor "test/internal/qemu"
	qemuimg "test/internal/qemu/image"
	"test/model"
	"test/nmap"
	"time"

	"github.com/digitalocean/go-qemu/qemu"
	"gopkg.in/yaml.v3"
)


const (
	// domain = "sontay.thinkmay.net"
	domain = ""
)


func init() {
	if domain == "" {
		return
	}


	result,err := exec.Command("sudo",
		"snap","install","certbot",
		"--classic").Output()
	fmt.Println("---------------------------")
	fmt.Println("setting up certbot")
	fmt.Println(string(result))
	fmt.Println("---------------------------")
	if err != nil {
		fmt.Printf("%s\n",err.Error())
	}


	fmt.Println("---------------------------")
	fmt.Println("setting up ssl certificate")
	result,err = exec.Command("sudo",
		"certbot","certonly", "--standalone",
		"--preferred-challenges","http",
		"-d",domain,
		"-m","huyhoangdo0205@gmail.com",
		"--agree-tos","-n").Output()
	if err != nil {
		fmt.Printf("%s\n",err.Error())
	}

	fmt.Println(string(result))
	fmt.Println("---------------------------")
}

type AuthHeader struct {
	APIKey 		*string `json:"api_key"`
	APIToken 	*string `json:"api_token"`
}

func (auth *AuthHeader)ParseReq(r *http.Request) {
	headers := map[string]string{}
	for k, v := range r.Header {
		headers[k] = v[0]
	}

	data,_ := json.Marshal(headers)
	json.Unmarshal(data, auth)
}


type VirtDaemon struct {
	APIKeys map[string]string
	hypervisor *qemuhypervisor.QEMUHypervisor
	libvirt *libvirt.Libvirt
}

func NewVirtDaemon(port int) *VirtDaemon{
	daemon := &VirtDaemon{
		APIKeys: map[string]string{
			"iuvgb2qg7rwyashbvkaiueg2v3uqfwaivusgfvy" : "972gavszdufg8oywfabsdzvoaiwgefb",
		},
		hypervisor: qemuhypervisor.NewQEMUHypervisor(),
		libvirt: libvirt.NewLibvirt(),
	}

	http.HandleFunc("/deploy", 		daemon.deployVM)
	http.HandleFunc("/start", 		daemon.startVM)
	http.HandleFunc("/stop", 		daemon.stopVM)
	http.HandleFunc("/delete", 		daemon.deleteVM)
	http.HandleFunc("/status", 		daemon.statusVM)

	http.HandleFunc("/vms", 		daemon.listVMs)

	http.HandleFunc("/disks", 		daemon.listDisks)
	http.HandleFunc("/disk/clone", 	daemon.cloneDisk)
	http.HandleFunc("/disk/delete", daemon.cloneDisk)

	http.HandleFunc("/gpus", 		daemon.listGPUs)
	http.HandleFunc("/ifaces", 		daemon.listIfaces)


	go func ()  {
		if domain == "" {
			err := http.ListenAndServe("0.0.0.0:8090", nil)
			if err != nil {
				panic(err)
			}
		} else {
			certFile := fmt.Sprintf("/etc/letsencrypt/live/%s/fullchain.pem",domain)
			keyFile  := fmt.Sprintf("/etc/letsencrypt/live/%s/privkey.pem",  domain)
			err := http.ListenAndServeTLS("0.0.0.0:443", certFile,keyFile, nil)
			if err != nil {
				panic(err)
			}
		}
	}()
	return daemon
}








func (daemon *VirtDaemon)deployVM(w http.ResponseWriter, r *http.Request) {
	body,err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}


	server := struct{
		VCPU int `yaml:"vcpus"`
		RAM  int `yaml:"ram"`

		GPU []model.GPU `yaml:"gpu"`
		Volume []model.Volume`yaml:"volume"`
		Interface []model.Iface`yaml:"interface"`
	}{ }

	err = yaml.Unmarshal(body,&server)
	if err != nil {
		w.WriteHeader(400)
		io.WriteString(w, "invalid yaml")
		return
	}

	name,err := daemon.libvirt.CreateVM(
		server.VCPU,
		server.RAM,
		server.GPU,
		server.Volume,
		server.Interface,
	)
	if err != nil {
		w.WriteHeader(400)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(200)
	io.WriteString(w, name)
}

func (daemon *VirtDaemon)stopVM(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)
	body,err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}

	found := false
	doms := daemon.hypervisor.ListDomain()
	for _, d := range doms {
		if d.Name == string(body) && d.Status == qemu.StatusRunning {
			found = true 
		}
	}

	if !found {
		w.WriteHeader(404)
		return
	}


	err = daemon.libvirt.StopVM(string(body))
	if err != nil {
		w.WriteHeader(400)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Println("stopped VM")
}
func (daemon *VirtDaemon)startVM(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)
	body,err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}

	found := false
	doms := daemon.hypervisor.ListDomain()
	for _, d := range doms {
		if d.Name == string(body) && d.Status == qemu.StatusShutdown {
			found = true
		}
	}
	if !found {
		w.WriteHeader(404)
		return
	}

	w.WriteHeader(200)
	fmt.Println("started VM")
}
func (daemon *VirtDaemon)deleteVM(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)
	body,err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}

	found := false
	doms := daemon.hypervisor.ListDomain()
	for _, d := range doms {
		if d.Name == string(body) {
			found = true
		}
	}
	if !found {
		w.WriteHeader(404)
		return
	}

	err = daemon.libvirt.DeleteVM(string(body))
	if err != nil {
		w.WriteHeader(400)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(200)
	fmt.Println("deleted VM")
}


func (daemon *VirtDaemon)statusVM(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)

	body,err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}



	if err != nil {
		w.WriteHeader(200)
		io.WriteString(w, err.Error())
		return
	}

	doms := daemon.hypervisor.ListDomain()
	for _, d := range doms {
		if d.Name == string(body) {
			w.WriteHeader(200)
			io.WriteString(w, d.Status.String())
			return
		}
	}
	

	w.WriteHeader(404)
	io.WriteString(w, "VM not found")
}


func (daemon *VirtDaemon)listVMs(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)



	doms    := daemon.libvirt.ListDomains()
	qemudom := daemon.hypervisor.ListDomain()
	iface   := daemon.libvirt.ListIfaces()

	result := map[string][]model.Domain{}
	networks := nmap.FindIPMac("192.168.1.*")

	for _, d := range qemudom {
		for _, d2 := range doms {
			if d.Name == *d2.Name {
				macs := []string{}
				for _, i2 := range d2.Interfaces {
					for _, i3 := range iface {
						if i2.Target == nil {
							continue
						}

						if i3.Name == i2.Target.Dev {
							macs = append(macs, *i3.Mac.Address)
						}
					}
				}

				ips := []string{}
				for k, v := range networks {
					for _, v2 := range macs {
						if strings.ToLower(v2) == strings.ToLower(k) {
							ips = append(ips, v)
						}
					}
				}

				d2.PrivateIP = &ips
				if result[d.Status.String()] == nil {
					result[d.Status.String()] = []model.Domain{d2}
				} else {
					result[d.Status.String()] = append(result[d.Status.String()],d2)
				}
			}
		}
	}
	data,_ := yaml.Marshal(result)

	w.WriteHeader(200)
	io.WriteString(w, string(data))
}

func (daemon *VirtDaemon)listDisks(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)



	volume := daemon.libvirt.ListDisks()
	result := struct{
		Active []model.Volume `yaml:"active"`
		Available []model.Volume `yaml:"open"`
	}{
		Active: volume,
		Available: []model.Volume{},
	}

	qemudom := daemon.hypervisor.ListDomain()
	for _, v := range volume {
		add := true
		for _, d := range qemudom {
			for _, bd := range d.BlockDevs {
				if bd.Inserted.File == v.Path {
					add = false
				
				}
			}
		}
		if add {
			result.Available = append(result.Available, v)
		}
	}
	w.WriteHeader(200)
	data,_ := yaml.Marshal(result)
	io.WriteString(w, string(data))
}
func (daemon *VirtDaemon)cloneDisk(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)


	body,_ := io.ReadAll(r.Body)
	in := struct{ 
		Source model.Volume `yaml:"source"` 
		Size int `yaml:"size"` 
	}{}
	err := yaml.Unmarshal(body,&in)
	if err != nil {
		w.WriteHeader(400)
		io.WriteString(w, "invalid body")
	}


	name := fmt.Sprintf("/disk/2TB1/autogen/%d.qcow2", time.Now().Nanosecond())
	err = qemuimg.CloneVolume(in.Source.Path,name,in.Size)
	if err != nil {
		w.WriteHeader(400)
		io.WriteString(w, err.Error())
	}

	volume := daemon.libvirt.ListDisks()
	for _,v := range volume {
		if v.Path == name {
			w.WriteHeader(200)
			data,_ := yaml.Marshal(v)
			io.WriteString(w, string(data))
			return
		}
	}

	w.WriteHeader(400)
	io.WriteString(w, "failed")
}
func (daemon *VirtDaemon)listIfaces(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)


	

	ifaces := daemon.libvirt.ListIfaces()
	result := struct{
		Active []model.Iface `yaml:"active"`
		Available []model.Iface `yaml:"open"`
	}{
		Active: ifaces,
		Available: []model.Iface {},
	}

	qemudom := daemon.libvirt.ListDomains()
	for _, v := range ifaces {
		add := true
		for _, d := range qemudom {
			for _, bd := range d.Interfaces {
				if bd.Source.Dev == v.Name ||
				   v.Type != "ethernet" {
					add = false
				} else if bd.Target == nil {
				} else if bd.Target.Dev == v.Name {
					add = false
				}
			}
		}
		if add {
			result.Available = append(result.Available, v)
		}
	}
	w.WriteHeader(200)
	data,_ := yaml.Marshal(result)
	io.WriteString(w, string(data))
}

func (daemon *VirtDaemon)listGPUs(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)



	filtered := []model.GPU{}
	gpus 	:= daemon.libvirt.ListGPUs()
	domains := daemon.libvirt.ListDomains()
	qemudom := daemon.hypervisor.ListDomain()

	for _, g := range gpus {
		add := true
		for _, d := range domains {

			ignore := false
			for _, d2 := range qemudom {
				if *d.Name == d2.Name && 
					d2.Status != qemu.StatusRunning &&
					d2.Status != qemu.StatusPaused {
					ignore = true
				}
			}
			
			if ignore {
				continue
			}


			for _, hd := range d.Hostdevs {
				for _, v := range g.Capability.IommuGroup.Address {
					if  hd.SourceAddress.Bus == v.Bus &&
						hd.SourceAddress.Domain == v.Domain && 
						hd.SourceAddress.Function == v.Function && 
						hd.SourceAddress.Slot == v.Slot{
							add = false
					}
				}
			}
		}

		if !add {
			continue
		}

		filtered = append(filtered, g)
	}

	data,_ := yaml.Marshal(struct{
		Active []model.GPU `yaml:"active"`
		Available []model.GPU `yaml:"open"`
	}{
		Active: gpus,
		Available: filtered,
	})

	w.WriteHeader(200)
	io.WriteString(w, string(data))
}