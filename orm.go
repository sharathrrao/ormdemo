package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"orm/configuration"
	"orm/dynamicrbac"

	prompt "github.com/c-bata/go-prompt"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp""
)

// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

func completer(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
		{Text: "groups", Description: "Combine users with specific rules"},
	}
	return prompt.FilterHasPrefix(s, in.GetWordBeforeCursor(), true)
}

func createGroup(group string, config configuration.Cluster) {
	response, err := http.Post("http://"+config.Restserverip+":"+config.Restserverport+"/keystone/group/"+group, "application/text", bytes.NewReader([]byte(group)))
	if err != nil {
		fmt.Println("Http Request Failed ", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
}
func addUserToGroup(user string, group string, config configuration.Cluster) {
	response, err := http.Post("http://"+config.Restserverip+":"+config.Restserverport+"/keystone/adduser/"+user+"/group/"+group, "application/text", bytes.NewReader([]byte(user)))
	if err != nil {
		fmt.Println("Http Request Failed ", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
}

func clipromptRoles(config configuration.Cluster) {
	fmt.Println(" Please select from the following Options")
	fmt.Println(" 1. Create Role")
	fmt.Println(" 2. List Role")
	fmt.Println(" 3. Delete Role")
	fmt.Println(" 4. Add Role to Group")
	selection := prompt.Input("> ", completer)
	//fmt.Println("You selected " + t)
	switch selection {
	case "1":
		fmt.Println("Enter Role Name")
		roleName := prompt.Input(">", completer)
		Createrole(roleName, config)
	case "2":
		listrole(config)
	case "4":
		fmt.Println("Enter Role")
		role := prompt.Input(">", completer)
		fmt.Println("Enter Group for the role ")
		group := prompt.Input(">", completer)
		addRoleToGroup(role, group, config)
	}
}

func main() {
	configurationData := configuration.LoadConfig("/etc/demo.cfg")
	//fmt.Println(clusterdata)
	fmt.Println(" Loading Configuration Data.....")

	fmt.Println("Identified", len(configurationData.Clusters), "OpenStack Deployments ")
	fmt.Println()

	cluster1 := configurationData.Clusters[1]
	clipromptRapidRbac(cluster1)

	//TestChangePolicy()

}

func addRoleToGroup(role string, group string, config configuration.Cluster) {
	response, err := http.Post("http://"+config.Restserverip+":"+config.Restserverport+"/keystone/addrole/"+role+"/group/"+group, "application/text", bytes.NewReader([]byte(role)))
	if err != nil {
		fmt.Println("Http Request Failed ", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		if role != "_member_" {
			fmt.Println(string(data))
		}
	}
}

//Createrole - Create a new role role
func Createrole(role string, config configuration.Cluster) {
	response, err := http.Post("http://"+config.Restserverip+":"+config.Restserverport+"/keystone/role/"+role, "application/text", bytes.NewReader([]byte(role)))
	if err != nil {
		fmt.Println("Http Request Failed ", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}

}
func listrole(config configuration.Cluster) {
	response, err := http.Get("http://" + config.Restserverip + ":" + config.Restserverport + "/keystone/role")
	if err != nil {
		log.Fatal("HTTP Get Request Failed", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}

}

func listgroups(config configuration.Cluster) {
	response, err := http.Get("http://" + config.Restserverip + ":" + config.Restserverport + "/keystone/group")
	if err != nil {
		log.Fatal("HTTP Get Request Failed", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
}

//Createuser create a new user
func Createuser(user string, config configuration.Cluster) {
	response, err := http.Post("http://"+config.Restserverip+":"+config.Restserverport+"/keystone/user/"+user, "application/text", bytes.NewReader([]byte(user)))
	if err != nil {
		fmt.Println("Http Request Failed ", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
}

//Assignusertorole - assings a user to a role
func Assignusertorole(user string, role string) {

}
func clipromptRapidRbac(config configuration.Cluster) {
	var exitloop bool
	exitloop = false
	for i := 0; exitloop == false; i++ {
		fmt.Println(" Please select from the following Options")
		fmt.Println(" 1. Create Group")
		fmt.Println(" 2. Create User")
		fmt.Println(" 3. Add Role To Group")
		fmt.Println(" 4. Dynamic RBAC - Add Role to Policy")
		//fmt.Println(" 4. Create User")
		selection := prompt.Input("> ", completer)
		switch selection {
		case "1":
			fmt.Println("Enter Name of Group")
			groupName := prompt.Input(">", completer)
			createGroup(groupName, config)
		case "2":
			fmt.Println(" Enter Name of User")
			user := prompt.Input(">", completer)
			fmt.Println("Enter Group for the user ")
			group := prompt.Input(">", completer)
			Createuser(user, config)
			addUserToGroup(user, group, config)
			addRoleToGroup("_member_", group, config)
		case "3":
			fmt.Println("Enter Role")
			role := prompt.Input(">", completer)
			fmt.Println("Enter Group for the role ")
			group := prompt.Input(">", completer)
			addRoleToGroup(role, group, config)
		case "4":
			fmt.Println("Enter API To Add")
			api := prompt.Input(">", completer)
			fmt.Println("Enter Role to Add")
			role := prompt.Input(">", completer)
			dynamicrbac.UpdateConfigMap(api, role)
			//UpdateConfigMap(api, role)
		case "5":
			fmt.Println("Enter User name")
			user := prompt.Input(">", completer)
			Createuser(user, config)
		case "6":
			listgroups(config)
		default:
			exitloop = true
		}
	}
}
