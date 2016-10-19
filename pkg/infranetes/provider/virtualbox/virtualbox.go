package virtualbox

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/sjpotter/infranetes/pkg/infranetes/provider"
	"github.com/sjpotter/infranetes/pkg/infranetes/provider/common"

	"github.com/apcera/libretto/virtualmachine/virtualbox"

	kubeapi "k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

type vboxProvider struct {
	netDevice string
	vmSrc     string
}

func init() {
	provider.PodProviders.RegisterProvider("virtualbox", NewVBoxProvider)
}

type vboxConfig struct {
	NetDevice string
	VMSrc     string
}

func NewVBoxProvider() (provider.PodProvider, error) {
	var conf vboxConfig

	file, err := ioutil.ReadFile("virtualbox.json")
	if err != nil {
		return nil, fmt.Errorf("File error: %v\n", err)
	}

	json.Unmarshal(file, &conf)

	return &vboxProvider{
		netDevice: conf.NetDevice,
		vmSrc:     conf.VMSrc,
	}, nil
}

func (*vboxProvider) UpdatePodState(cPodData *common.PodData) {
	cPodData.UpdatePodState()
}

func (v *vboxProvider) RunPodSandbox(req *kubeapi.RunPodSandboxRequest) (*common.PodData, error) {
	config := virtualbox.Config{
		NICs: []virtualbox.NIC{
			{Idx: 1, Backing: virtualbox.Bridged, BackingDevice: v.netDevice},
		},
	}

	vm := &virtualbox.VM{Src: v.vmSrc,
		Config: config,
	}

	if err := vm.Provision(); err != nil {
		return nil, fmt.Errorf("Failed to Provision: %v", err)
	}

	ips, err := vm.GetIPs()
	if err != nil {
		return nil, fmt.Errorf("CreatePodSandbox: error in GetIPs(): %v", err)
	}

	ip := ips[0].String()

	client, err := common.CreateClient(ip)
	if err != nil {
		return nil, fmt.Errorf("CreatePodSandbox: error in createClient(): %v", err)
	}

	name := vm.GetName()

	podData := common.NewPodData(vm, &name, req.Config.Metadata, req.Config.Annotations, req.Config.Labels, ip, client, nil)

	return podData, nil
}

func (v *vboxProvider) StopPodSandbox(podData *common.PodData) {
}

func (v *vboxProvider) RemovePodSandbox(podData *common.PodData) {
}

func (v *vboxProvider) PodSandboxStatus(podData *common.PodData) {
}
