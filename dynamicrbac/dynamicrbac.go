package dynamicrbac

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		//fmt.Printf("Home is %s", h)
		return h
	}

	return os.Getenv("USERPROFILE") // windows
}

//UpdateConfigMap - update the config map
func UpdateConfigMap(api string, role string) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// bootstrap config
	// /fmt.Println()
	//fmt.Println("Using kubeconfig: ", *kubeconfig)
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	/*pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		panic(err.Error())
	}
	for _, pod := range pods.Items {
		if pod.Status.Phase == "Running" {
			//			fmt.Println(pod.Name, "\t", pod.Status.Phase)

		} //else {
		//fmt.Println(pod.Name+" is NOT RUNNING");
		//}

	}*/

	//configmaps, err := clientset.CoreV1().ConfigMaps("openstack").List(metav1.ListOptions{})
	//if err != nil {
	//	panic(err.Error())
	//}
	time.Sleep(10 * time.Second)
	novaConfigMap, err := clientset.CoreV1().ConfigMaps("openstack").Get("nova-etc", metav1.GetOptions{})
	novaConfigMapData := novaConfigMap.Data
	/*for key := range novaConfigMapData {
		fmt.Println("Key:", key) //, "Value:", value)
	}*/

	//fmt.Println(novaPolicyData)
	//policyStrings := strings.Split(novaPolicyData, ":")
	//var originalPolicy string

	//var novaPolicyData1 = originalPolicy //novaPolicyData + "" + "os_compute_api:os-aggregates:show: rule:admin_support\n"
	novaPolicyData := novaConfigMapData["policy.yaml"]
	addToNovaPolicy(&novaPolicyData, role)
	//fmt.Println("POLICY DATA POST UPDATE is ", novaPolicyData)
	//novaPolicyData1 := novaPolicyData + "\n" + "    os_compute_api:os-aggregates:show: rule:admin_support\n"
	novaConfigMapData["policy.yaml"] = novaPolicyData

	novaConfigMap.Data = novaConfigMapData
	configmapclient, err := clientset.CoreV1().ConfigMaps("openstack").Update(novaConfigMap)
	if err != nil {
		panic(err.Error())
	}
	log.Output(2, configmapclient.Name)
	listpods, err := clientset.CoreV1().Pods("openstack").List(metav1.ListOptions{})
	if err != nil {
		fmt.Println(err)
	}

	// /deletePolicy := metav1.DeletePropagationBackground
	for _, d := range listpods.Items {
		if strings.HasPrefix(d.Name, "nova") {
			err := clientset.CoreV1().Pods("openstack").Delete(d.Name, nil)
			if err != nil {
				fmt.Println("error", err)
			}
		}

	}
	time.Sleep(20 * time.Second)
	podsRunning := false
	for i := 0; podsRunning == false; i++ {
		listpods, err := clientset.CoreV1().Pods("openstack").List(metav1.ListOptions{})
		if err != nil {
			fmt.Println(err)
			podsRunning = true
		}

		podsRunning = true
		for _, pd := range listpods.Items {
			if strings.HasPrefix(pd.Name, "nova") {
				if pd.Status.Phase != "Running" {
					podsRunning = false

				}
			}
		}
		time.Sleep(10 * time.Second)
	}

	//fmt.Println(configmapclient.Name)
	fmt.Println("Successfully added Role to Policy")
	//fmt.Println(policyStrings[0])

}

func addToNovaPolicy(policydata *string, role string) string {
	testData := *policydata
	testDataArray := strings.Split(testData, "\n")
	var newArray []string
	newArray = make([]string, len(testDataArray))
	addRole := false
	for i := 0; i < len(testDataArray); i++ {
		currentLine := strings.Split(testDataArray[i], ":")
		if strings.TrimSpace(currentLine[0]) == "os_compute_api" {
			if strings.TrimSpace(currentLine[1]) == "os-aggregates" {
				//if strings.TrimSpace(currentLine[2]) == "create" || strings.TrimSpace(currentLine[2]) == "add_host" || strings.TrimSpace(currentLine[2]) == "index" || strings.TrimSpace(currentLine[2]) == "show" {
				if strings.TrimSpace(currentLine[2]) == "delete" {
					addRole = true
					//currentLine[3] += "role:tenant_nova_create"
					//fmt.Println(currentLine)
				}
			}
		}
		newArray[i] = strings.Join(currentLine, ":")
		if addRole {
			newArray[i] += " or role:" + role
			addRole = false
		}
	}
	testData = strings.Join(newArray, "\n")
	*policydata = testData
	return *policydata
}

/*func main() {
	updateConfigMap()
}*/

//TestChangePolicy test function
func TestChangePolicy() {

	var test = "admin_api: is_admin:True\n" +
		"admin_or_owner: is_admin:True or project_id:%(project_id)s\n" +
		"cells_scheduler_filter:DifferentCellFilter: is_admin:True\n    cells_scheduler_filter:TargetCellFilter: is_admin:True\n    context_is_admin: role:admin\n    network:attach_external_network: is_admin:True\n    os_compute_api:extension_info:discoverable: '@'\n    os_compute_api:extensions: rule:admin_or_owner\n    os_compute_api:extensions:discoverable: '@'\n    os_compute_api:flavors: rule:admin_or_owner\n    os_compute_api:flavors:discoverable: '@'\n    os_compute_api:image-metadata:discoverable: '@'\n    os_compute_api:image-size: rule:admin_or_owner\n    os_compute_api:image-size:discoverable: '@'\n    os_compute_api:images:discoverable: '@'\n    os_compute_api:ips:discoverable: '@'\n    os_compute_api:ips:index: rule:admin_or_owner\n    os_compute_api:ips:show: rule:admin_or_owner\n    os_compute_api:limits: rule:admin_or_owner\n    os_compute_api:limits:discoverable: '@'\n    os_compute_api:os-admin-actions: rule:admin_api\n    os_compute_api:os-admin-actions:discoverable: '@'\n    os_compute_api:os-admin-actions:inject_network_info: rule:admin_api\n    os_compute_api:os-admin-actions:reset_network: rule:admin_api\n    os_compute_api:os-admin-actions:reset_state: rule:admin_api\n    os_compute_api:os-admin-password: rule:admin_or_owner\n    os_compute_api:os-admin-password:discoverable: '@'\n    os_compute_api:os-agents: rule:admin_api\n    os_compute_api:os-agents:discoverable: '@'\n    os_compute_api:os-aggregates:add_host: rule:admin_api\n    os_compute_api:os-aggregates:create: rule:admin_api\n    os_compute_api:os-aggregates:delete: rule:admin_api\n    os_compute_api:os-aggregates:discoverable: '@'\n    os_compute_api:os-aggregates:index: rule:admin_api\n    os_compute_api:os-aggregates:remove_host: rule:admin_api\n    os_compute_api:os-aggregates:set_metadata: rule:admin_api\n    os_compute_api:os-aggregates:show: rule:admin_api\n    os_compute_api:os-aggregates:update: rule:admin_api\n    os_compute_api:os-assisted-volume-snapshots:create: rule:admin_api\n    os_compute_api:os-assisted-volume-snapshots:delete: rule:admin_api\n    os_compute_api:os-assisted-volume-snapshots:discoverable: '@'\n    os_compute_api:os-attach-interfaces: rule:admin_or_owner\n    os_compute_api:os-attach-interfaces:create: rule:admin_or_owner\n    os_compute_api:os-attach-interfaces:delete: rule:admin_or_owner\n    os_compute_api:os-attach-interfaces:discoverable: '@'\n    os_compute_api:os-availability-zone:detail: rule:admin_api\n    os_compute_api:os-availability-zone:discoverable: '@'\n    os_compute_api:os-availability-zone:list: rule:admin_or_owner\n    os_compute_api:os-baremetal-nodes: rule:admin_api\n    os_compute_api:os-baremetal-nodes:discoverable: '@'\n    os_compute_api:os-block-device-mapping-v1:discoverable: '@'\n    os_compute_api:os-block-device-mapping:discoverable: '@'\n    os_compute_api:os-cells: rule:admin_api\n    os_compute_api:os-cells:create: rule:admin_api\n    os_compute_api:os-cells:delete: rule:admin_api\n    os_compute_api:os-cells:discoverable: '@'\n    os_compute_api:os-cells:sync_instances: rule:admin_api\n    os_compute_api:os-cells:update: rule:admin_api\n    os_compute_api:os-certificates:create: rule:admin_or_owner\n    os_compute_api:os-certificates:discoverable: '@'\n    os_compute_api:os-certificates:show: rule:admin_or_owner\n    os_compute_api:os-cloudpipe: rule:admin_api\n    os_compute_api:os-cloudpipe:discoverable: '@'\n    os_compute_api:os-config-drive: rule:admin_or_owner\n    os_compute_api:os-config-drive:discoverable: '@'\n    os_compute_api:os-console-auth-tokens: rule:admin_api\n    os_compute_api:os-console-auth-tokens:discoverable: '@'\n    os_compute_api:os-console-output: rule:admin_or_owner\n    os_compute_api:os-console-output:discoverable: '@'\n    os_compute_api:os-consoles:create: rule:admin_or_owner\n    os_compute_api:os-consoles:delete: rule:admin_or_owner\n    os_compute_api:os-consoles:discoverable: '@'\n    os_compute_api:os-consoles:index: rule:admin_or_owner\n    os_compute_api:os-consoles:show: rule:admin_or_owner\n    os_compute_api:os-create-backup: rule:admin_or_owner\n    os_compute_api:os-create-backup:discoverable: '@'\n    os_compute_api:os-deferred-delete: rule:admin_or_owner\n    os_compute_api:os-deferred-delete:discoverable: '@'\n    os_compute_api:os-evacuate: rule:admin_api\n    os_compute_api:os-evacuate:discoverable: '@'\n    os_compute_api:os-extended-availability-zone: rule:admin_or_owner\n    os_compute_api:os-extended-availability-zone:discoverable: '@'\n    os_compute_api:os-extended-server-attributes: rule:admin_api\n    os_compute_api:os-extended-server-attributes:discoverable: '@'\n    os_compute_api:os-extended-status: rule:admin_or_owner\n    os_compute_api:os-extended-status:discoverable: '@'\n    os_compute_api:os-extended-volumes: rule:admin_or_owner\n    os_compute_api:os-extended-volumes:discoverable: '@'\n    os_compute_api:os-fixed-ips: rule:admin_api\n    os_compute_api:os-fixed-ips:discoverable: '@'\n    os_compute_api:os-flavor-access: rule:admin_or_owner\n    os_compute_api:os-flavor-access:add_tenant_access: rule:admin_api\n    os_compute_api:os-flavor-access:discoverable: '@'\n    os_compute_api:os-flavor-access:remove_tenant_access: rule:admin_api\n    os_compute_api:os-flavor-extra-specs:create: rule:admin_api\n    os_compute_api:os-flavor-extra-specs:delete: rule:admin_api\n    os_compute_api:os-flavor-extra-specs:discoverable: '@'\n    os_compute_api:os-flavor-extra-specs:index: rule:admin_or_owner\n    os_compute_api:os-flavor-extra-specs:show: rule:admin_or_owner\n    os_compute_api:os-flavor-extra-specs:update: rule:admin_api\n    os_compute_api:os-flavor-manage: rule:admin_api\n    os_compute_api:os-flavor-manage:discoverable: '@'\n    os_compute_api:os-flavor-rxtx: rule:admin_or_owner\n    os_compute_api:os-flavor-rxtx:discoverable: '@'\n    os_compute_api:os-floating-ip-dns: rule:admin_or_owner\n    os_compute_api:os-floating-ip-dns:discoverable: '@'\n    os_compute_api:os-floating-ip-dns:domain:delete: rule:admin_api\n    os_compute_api:os-floating-ip-dns:domain:update: rule:admin_api\n    os_compute_api:os-floating-ip-pools: rule:admin_or_owner\n    os_compute_api:os-floating-ip-pools:discoverable: '@'\n    os_compute_api:os-floating-ips: rule:admin_or_owner\n    os_compute_api:os-floating-ips-bulk: rule:admin_api\n    os_compute_api:os-floating-ips-bulk:discoverable: '@'\n    os_compute_api:os-floating-ips:discoverable: '@'\n    os_compute_api:os-fping: rule:admin_or_owner\n    os_compute_api:os-fping:all_tenants: rule:admin_api\n    os_compute_api:os-fping:discoverable: '@'\n    os_compute_api:os-hide-server-addresses: is_admin:False\n    os_compute_api:os-hide-server-addresses:discoverable: '@'\n    os_compute_api:os-hosts: rule:admin_api\n    os_compute_api:os-hosts:discoverable: '@'\n    os_compute_api:os-hypervisors: rule:admin_api\n    os_compute_api:os-hypervisors:discoverable: '@'\n    os_compute_api:os-instance-actions: rule:admin_or_owner\n    os_compute_api:os-instance-actions:discoverable: '@'\n    os_compute_api:os-instance-actions:events: rule:admin_api\n    os_compute_api:os-instance-usage-audit-log: rule:admin_api\n    os_compute_api:os-instance-usage-audit-log:discoverable: '@'\n    os_compute_api:os-keypairs: rule:admin_or_owner\n    os_compute_api:os-keypairs:create: rule:admin_api or user_id:%(user_id)s\n    os_compute_api:os-keypairs:delete: rule:admin_api or user_id:%(user_id)s\n    os_compute_api:os-keypairs:discoverable: '@'\n    os_compute_api:os-keypairs:index: rule:admin_api or user_id:%(user_id)s\n    os_compute_api:os-keypairs:show: rule:admin_api or user_id:%(user_id)s\n    os_compute_api:os-lock-server:discoverable: '@'\n    os_compute_api:os-lock-server:lock: rule:admin_or_owner\n    os_compute_api:os-lock-server:unlock: rule:admin_or_owner\n    os_compute_api:os-lock-server:unlock:unlock_override: rule:admin_api\n    os_compute_api:os-migrate-server:discoverable: '@'\n    os_compute_api:os-migrate-server:migrate: rule:admin_api\n    os_compute_api:os-migrate-server:migrate_live: rule:admin_api\n    os_compute_api:os-migrations:discoverable: '@'\n    os_compute_api:os-migrations:index: rule:admin_api\n    os_compute_api:os-multinic: rule:admin_or_owner\n    os_compute_api:os-multinic:discoverable: '@'\n    os_compute_api:os-multiple-create:discoverable: '@'\n    os_compute_api:os-networks: rule:admin_api\n    os_compute_api:os-networks-associate: rule:admin_api\n    os_compute_api:os-networks-associate:discoverable: '@'\n    os_compute_api:os-networks:discoverable: '@'\n    os_compute_api:os-networks:view: rule:admin_or_owner\n    os_compute_api:os-pause-server:discoverable: '@'\n    os_compute_api:os-pause-server:pause: rule:admin_or_owner\n    os_compute_api:os-pause-server:unpause: rule:admin_or_owner\n    os_compute_api:os-pci:detail: rule:admin_api\n    os_compute_api:os-pci:discoverable: '@'\n    os_compute_api:os-pci:index: rule:admin_api\n    os_compute_api:os-pci:pci_servers: rule:admin_or_owner\n    os_compute_api:os-pci:show: rule:admin_api\n    os_compute_api:os-quota-class-sets:discoverable: '@'\n    os_compute_api:os-quota-class-sets:show: is_admin:True or quota_class:%(quota_class)s\n    os_compute_api:os-quota-class-sets:update: rule:admin_api\n    os_compute_api:os-quota-sets:defaults: '@'\n    os_compute_api:os-quota-sets:delete: rule:admin_api\n    os_compute_api:os-quota-sets:detail: rule:admin_api\n    os_compute_api:os-quota-sets:discoverable: '@'\n    os_compute_api:os-quota-sets:show: rule:admin_or_owner\n    os_compute_api:os-quota-sets:update: rule:admin_api\n    os_compute_api:os-remote-consoles: rule:admin_or_owner\n    os_compute_api:os-remote-consoles:discoverable: '@'\n    os_compute_api:os-rescue: rule:admin_or_owner\n    os_compute_api:os-rescue:discoverable: '@'\n    os_compute_api:os-scheduler-hints:discoverable: '@'\n    os_compute_api:os-security-group-default-rules: rule:admin_api\n    os_compute_api:os-security-group-default-rules:discoverable: '@'\n    os_compute_api:os-security-groups: rule:admin_or_owner\n    os_compute_api:os-security-groups:discoverable: '@'\n    os_compute_api:os-server-diagnostics: rule:admin_api\n    os_compute_api:os-server-diagnostics:discoverable: '@'\n    os_compute_api:os-server-external-events:create: rule:admin_api\n    os_compute_api:os-server-external-events:discoverable: '@'\n    os_compute_api:os-server-groups: rule:admin_or_owner\n    os_compute_api:os-server-groups:discoverable: '@'\n    os_compute_api:os-server-password: rule:admin_or_owner\n    os_compute_api:os-server-password:discoverable: '@'\n    os_compute_api:os-server-tags:delete: '@'\n    os_compute_api:os-server-tags:delete_all: '@'\n    os_compute_api:os-server-tags:discoverable: '@'\n    os_compute_api:os-server-tags:index: '@'\n    os_compute_api:os-server-tags:show: '@'\n    os_compute_api:os-server-tags:update: '@'\n    os_compute_api:os-server-tags:update_all: '@'\n    os_compute_api:os-server-usage: rule:admin_or_owner\n    os_compute_api:os-server-usage:discoverable: '@'\n    os_compute_api:os-services: rule:admin_api\n    os_compute_api:os-services:discoverable: '@'\n    os_compute_api:os-shelve:discoverable: '@'\n    os_compute_api:os-shelve:shelve: rule:admin_or_owner\n    os_compute_api:os-shelve:shelve_offload: rule:admin_api\n    os_compute_api:os-shelve:unshelve: rule:admin_or_owner\n    os_compute_api:os-simple-tenant-usage:discoverable: '@'\n    os_compute_api:os-simple-tenant-usage:list: rule:admin_api\n    os_compute_api:os-simple-tenant-usage:show: rule:admin_or_owner\n    os_compute_api:os-suspend-server:discoverable: '@'\n    os_compute_api:os-suspend-server:resume: rule:admin_or_owner\n    os_compute_api:os-suspend-server:suspend: rule:admin_or_owner\n    os_compute_api:os-tenant-networks: rule:admin_or_owner\n    os_compute_api:os-tenant-networks:discoverable: '@'\n    os_compute_api:os-used-limits: rule:admin_api\n    os_compute_api:os-used-limits:discoverable: '@'\n    os_compute_api:os-user-data:discoverable: '@'\n    os_compute_api:os-virtual-interfaces: rule:admin_or_owner\n    os_compute_api:os-virtual-interfaces:discoverable: '@'\n    os_compute_api:os-volumes: rule:admin_or_owner\n    os_compute_api:os-volumes-attachments:create: rule:admin_or_owner\n    os_compute_api:os-volumes-attachments:delete: rule:admin_or_owner\n    os_compute_api:os-volumes-attachments:discoverable: '@'\n    os_compute_api:os-volumes-attachments:index: rule:admin_or_owner\n    os_compute_api:os-volumes-attachments:show: rule:admin_or_owner\n    os_compute_api:os-volumes-attachments:update: rule:admin_api\n    os_compute_api:os-volumes:discoverable: '@'\n    os_compute_api:server-metadata:create: rule:admin_or_owner\n    os_compute_api:server-metadata:delete: rule:admin_or_owner\n    os_compute_api:server-metadata:discoverable: '@'\n    os_compute_api:server-metadata:index: rule:admin_or_owner\n    os_compute_api:server-metadata:show: rule:admin_or_owner\n    os_compute_api:server-metadata:update: rule:admin_or_owner\n    os_compute_api:server-metadata:update_all: rule:admin_or_owner\n    os_compute_api:server-migrations:discoverable: '@'\n    os_compute_api:servers:confirm_resize: rule:admin_or_owner\n    os_compute_api:servers:create: rule:admin_or_owner\n    os_compute_api:servers:create:attach_network: rule:admin_or_owner\n    os_compute_api:servers:create:attach_volume: rule:admin_or_owner\n    os_compute_api:servers:create:forced_host: rule:admin_api\n    os_compute_api:servers:create_image: rule:admin_or_owner\n    os_compute_api:servers:create_image:allow_volume_backed: rule:admin_or_owner\n    os_compute_api:servers:delete: rule:admin_or_owner\n    os_compute_api:servers:detail: rule:admin_or_owner\n    os_compute_api:servers:detail:get_all_tenants: rule:admin_api\n    os_compute_api:servers:discoverable: '@'\n    os_compute_api:servers:index: rule:admin_or_owner\n    os_compute_api:servers:index:get_all_tenants: rule:admin_api\n    os_compute_api:servers:migrations:delete: rule:admin_api\n    os_compute_api:servers:migrations:force_complete: rule:admin_api\n    os_compute_api:servers:migrations:index: rule:admin_api\n    os_compute_api:servers:migrations:show: rule:admin_api\n    os_compute_api:servers:reboot: rule:admin_or_owner\n    os_compute_api:servers:rebuild: rule:admin_or_owner\n    os_compute_api:servers:resize: rule:admin_or_owner\n    os_compute_api:servers:revert_resize: rule:admin_or_owner\n    os_compute_api:servers:show: rule:admin_or_owner\n    os_compute_api:servers:show:host_status: rule:admin_api\n    os_compute_api:servers:start: rule:admin_or_owner\n    os_compute_api:servers:stop: rule:admin_or_owner\n    os_compute_api:servers:trigger_crash_dump: rule:admin_or_owner\n    os_compute_api:servers:update: rule:admin_or_owner\n    os_compute_api:versions:discoverable: '@'"

	addToNovaPolicy(&test, "test")
	fmt.Println(test)

}
