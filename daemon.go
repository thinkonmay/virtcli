package virtdaemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"test/internal/libvirt"
	qemuhypervisor "test/internal/qemu"
	"test/model"

	"github.com/digitalocean/go-qemu/qemu"
	"gopkg.in/yaml.v3"
)

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

	http.HandleFunc("/deploy", 	daemon.deployVM)
	http.HandleFunc("/delete", 	daemon.deleteVM)
	http.HandleFunc("/status", 	daemon.statusVM)

	http.HandleFunc("/vms", 	daemon.listVMs)

	http.HandleFunc("/gpus", 	daemon.listGPUs)
	http.HandleFunc("/ifaces", 	daemon.listIfaces)
	http.HandleFunc("/disks", 	daemon.listDisks)

	go http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d",port), nil)
	return daemon
}



type VMDeployInfo struct {

}
func (inf *VMDeployInfo)ParseReq(r *http.Request) error {
	data,err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(data,inf)
}




func (daemon *VirtDaemon)deployVM(w http.ResponseWriter, r *http.Request) {
	inf := &VMDeployInfo{}
	inf.ParseReq(r)

	body,err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}


	server := struct{
	}{
	}


	err = yaml.Unmarshal(body,&server)
	if err != nil {
		w.WriteHeader(400)
		io.WriteString(w, "invalid yaml")
		return
	}

	// daemon.libvirt.CreateVM()

	w.WriteHeader(200)
	io.WriteString(w, "success")
}

func (daemon *VirtDaemon)deleteVM(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)

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


	server := struct{
		ServerID string `json:"server_id"`
	}{}
	err = json.Unmarshal(body,&server)
	if err != nil {
		w.WriteHeader(200)
		io.WriteString(w, err.Error())
		return
	}

	doms := daemon.hypervisor.ListDomain()
	for _, d := range doms {
		if d.Name == server.ServerID {
			w.WriteHeader(200)
			io.WriteString(w, d.Status.String())
			return
		}
	}
	

	w.WriteHeader(404)
	io.WriteString(w, "VM not found")
	return
}


func (daemon *VirtDaemon)listVMs(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)



	doms    := daemon.libvirt.ListDomains()
	qemudom := daemon.hypervisor.ListDomain()

	result := map[string][]model.Domain{}

	for _, d := range qemudom {
		for _, d2 := range doms {
			if d.Name == *d2.Name {
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
		Active []model.Volume `yaml:"available"`
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
		if !add {
			continue
		}

		result.Available = append(result.Active, v)
	}
	w.WriteHeader(200)
	data,_ := yaml.Marshal(result)
	io.WriteString(w, string(data))
}
func (daemon *VirtDaemon)listIfaces(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)



	daemon.libvirt.ListIfaces()
	w.WriteHeader(200)
	// io.WriteString(w, string(data))
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
		Active []model.GPU `json:"available"`
		Available []model.GPU `json:"open"`
	}{
		Active: gpus,
		Available: filtered,
	})

	w.WriteHeader(200)
	io.WriteString(w, string(data))
}