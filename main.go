package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	user "keycloack-transfer/users"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var config map[string]interface{}

func init() {
	log.SetPrefix("Record: ")
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Llongfile)
}
func CheckError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

var dbconf = readConfigFile()["db"].(map[string]interface{})
var dbUrl = "postgres://" + dbconf["user"].(string) + ":" + dbconf["password"].(string) + "@" + dbconf["host"].(string) + ":" + dbconf["port"].(string) + "/" + dbconf["database"].(string) + "?sslmode=disable"

func getUsers(limitt string, offset string) []user.User {
	userList := []user.User{}
	db, err := sql.Open("postgres", dbUrl)
	CheckError(err)
	CheckError(db.Ping())
	query, err := db.Query("select name,surname,username,email from tbl_user order by created_date limit " + limitt + " offset " + offset)
	defer db.Close()
	for query.Next() {
		newUser := user.User{}
		err = query.Scan(&newUser.Name, &newUser.Surname, &newUser.UserName, &newUser.Email)
		userList = append(userList, newUser)
	}
	return userList
}
func getUserCount() int {
	var count int
	queryString := "select count(distinct id) from tbl_user"
	db, err := sql.Open("postgres", dbUrl)
	CheckError(err)
	defer db.Close()
	row := db.QueryRow(queryString)
	newErr := row.Scan(&count)
	CheckError(newErr)
	return count

}

var jsonMap map[string]interface{}

func readConfigFile() map[string]interface{} {

	data, err := ioutil.ReadFile(os.Getenv("CONFIG_PATH"))

	CheckError(err)
	err = yaml.Unmarshal(data, &config)

	CheckError(err)
	return config
}

func main() {
	//recordLogs, _ := os.Create("records.log")
	//defer recordLogs.Close()
	//log.SetOutput(recordLogs)
	offset := 0
	for i := 0; i < getUserCount(); i++ {
		if i%dbconf["cursorSize"].(int) == 0 {
			for _, u := range getUsers(strconv.Itoa(dbconf["cursorSize"].(int)), strconv.Itoa(offset)) {
				newUser(u)
			}
			offset += dbconf["cursorSize"].(int)
		}
	}
	fmt.Println(1001 / 100)
}

func newUser(dbuser user.User) {
	keycloakConf := readConfigFile()["keycloak"].(map[string]interface{})
	registerUser := user.RegisterUser{}
	attribute := user.Attribute{}
	credential := user.Credential{}
	credential.Type = "password"
	credential.Value = "test123"
	attribute.Test = "tst"
	registerUser.EmailVerified = true
	registerUser.Username = dbuser.UserName
	registerUser.LastName = dbuser.Surname
	registerUser.FirstName = dbuser.Name
	registerUser.Email = dbuser.Email
	registerUser.Credentials = append(registerUser.Credentials, credential)
	registerUser.Attributes = attribute
	user, err := json.Marshal(registerUser)
	CheckError(err)

	client := &http.Client{}

	req, err := http.NewRequest("POST", keycloakConf["baseurl"].(string)+"auth/admin/realms/"+keycloakConf["realm"].(string)+"/users", bytes.NewBuffer(user))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+getToken())

	res, err := client.Do(req)
	CheckError(err)
	if res.StatusCode == 201 {
		log.Println("Save User Successfully")
		log.Println(bytes.NewBuffer(user))
	} else {
		log.Println(res.StatusCode)
	}

}

func getToken() string {
	keycloakConf := readConfigFile()["keycloak"].(map[string]interface{})
	data := strings.NewReader("grant_type=password&client_id=admin-cli&username=" + keycloakConf["user"].(string) + "&password=" + keycloakConf["password"].(string) + "&client_secret=" + keycloakConf["secretKey"].(string))
	res, err := http.Post(keycloakConf["baseurl"].(string)+"auth/realms/master/protocol/openid-connect/token", "application/x-www-form-urlencoded", data)
	CheckError(err)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	json.Unmarshal(body, &jsonMap)
	log.Println("Get Token is successfully")
	return jsonMap["access_token"].(string)
}
