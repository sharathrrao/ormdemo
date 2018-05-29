package configuration

import (
	"encoding/json"
	"fmt"

	"os"
)

//Config struct with restip
type Config struct {
	Clusters []Cluster `json:"cluster"`
}

//Cluster structure
type Cluster struct {
	ClusterID      string `json:"clusterid"`
	Restserverip   string `json:"restserverip"`
	Restserverport string `json:"restserverport"`
}

//LoadConfig - configuration file
func LoadConfig(file string) Config {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		fmt.Println("error is ", err)
	}
	// fmt.Println(config.Cluster.Restserverip)
	return config
}
func main() {
	configurationData := LoadConfig("/etc/demo.cfg")
	//clusterdata := configurationData.Cluster
	fmt.Println(configurationData)
}
