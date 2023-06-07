package virtdaemon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	qemuhypervisor "test/internal/qemu/hypervisor"
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
}

func NewVirtDaemon(port int) *VirtDaemon{
	daemon := &VirtDaemon{
		APIKeys: map[string]string{
			"iuvgb2qg7rwyashbvkaiueg2v3uqfwaivusgfvy" : "972gavszdufg8oywfabsdzvoaiwgefb",
		},
		hypervisor: qemuhypervisor.NewQEMUHypervisor(),
	}

	http.HandleFunc("/", 		daemon.deployVM)
	http.HandleFunc("/list", 	daemon.listVMs)

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


}





func (daemon *VirtDaemon)listVMs(w http.ResponseWriter, r *http.Request) {
	auth := &AuthHeader{}
	auth.ParseReq(r)



	doms := daemon.hypervisor.ListDomain()
	data,_ := json.MarshalIndent(doms,"","  ")

	w.WriteHeader(200)
	io.WriteString(w, string(data))
}