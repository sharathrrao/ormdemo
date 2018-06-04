package main

import (
	"fmt"
	"log"
	"net/http"

	sh "github.com/codeskyblue/go-sh"
	"github.com/gorilla/mux"
)

//CreateKeystoneRole - Command to create keystone role
func CreateKeystoneRole(role string) []byte {
	out, err := sh.Command("/opt/openstackcommand.sh", "role create", role).Output()
	if err != nil {
		log.Fatal(err)

	}
	return out

}

//ListRoles - command to list roles
func ListRoles(w http.ResponseWriter, r *http.Request) {
	out, err := sh.Command("/opt/openstackcommand.sh", "role list").Output()
	if err != nil {
		log.Fatal(err)
	}
	w.Write(out)

}

func listgroups(w http.ResponseWriter, r *http.Request) {
	out, err := sh.Command("/opt/openstackcommand.sh", "group list").Output()
	if err != nil {
		log.Fatal(err)
	}
	w.Write(out)

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/keystone/role/{role}", CreateRole).Methods("POST")
	router.HandleFunc("/keystone/{role}/user/{user}", AssignRole).Methods("POST")
	router.HandleFunc("/keystone/role", ListRoles).Methods("GET")
	router.HandleFunc("/keystone/addrole/{role}/group/{group}", addRoleToGroup).Methods("POST")
	router.HandleFunc("/keystone/group/{group}", createGroup).Methods("POST")
	router.HandleFunc("/keystone/adduser/{user}/group/{group}", addUserToGroup).Methods("POST")
	router.HandleFunc("/keystone/user/{user}", createUser).Methods("POST")
	router.HandleFunc("/keystone/groups", listgroups).Methods("GET")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func createUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]

	out, err := sh.Command("/opt/openstackcommand.sh", "user create --password password1 --project testproject", user).Output()
	if err != nil {
		log.Fatal(err)

	}
	log.Output(2, string(out))

	message := "Successfully Create a new User " + user
	w.Write([]byte(message))
}

func addUserToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user := vars["user"]
	group := vars["group"]
	out, err := sh.Command("/opt/openstackcommand.sh", "group add user ", group, user).Output()
	if err != nil {
		log.Fatal(err)

	}
	message := "Successfully added User " + user + " to Group " + group
	log.Output(2, string(out))
	w.Write([]byte(message))

}

func createGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	group := vars["group"]

	out, err := sh.Command("/home/esharao/openstackcommand.sh", "group create --domain default ", group).Output()
	if err != nil {
		log.Fatal(err)

	}
	log.Output(2, string(out))
	message := "Successfully Created Group " + group
	w.Write([]byte(message))
}

func addRoleToGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	role := vars["role"]
	group := vars["group"]
	out, err := sh.Command("/opt/openstackcommand.sh", "role add --project testproject --group "+group, role).Output()
	if err != nil {
		log.Fatal(err)

	}
	log.Output(2, string(out))
	message := "Successfully Added Role " + role + " to group " + group
	w.Write([]byte(message))

}

//CreateRole - create a role
func CreateRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	role := vars["role"]
	fmt.Println("Role Received is ", role)
	output := CreateKeystoneRole(role)
	w.Write(output)
}

//AssignRole assign a role tothe user
func AssignRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	role := vars["role"]
	user := vars["user"]
	fmt.Println("Role is ", role, " user is ", user)

}
