package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/term"
	"net/http"
	"os"
	//"syscall"
)

/*Contents: This package contains all functions related to authentication. They fall into 3 categories:
1. Usable
2. Not yet implemented
3. Unusable due restrictions on who can use the API during Beta
*/

//////////////////////| Usable auth functions | Login & Token refresh |///////////////////////////

type loginData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthTokens struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	RTDBToken    string `json:"rtbdToken"`
}

type AuthResponse struct {
	Data AuthTokens `json:"data"`
}

func Login(url string) AuthTokens { //client http.Client,
	var email string
	//var password string
	email = os.Getenv("cyberspace_email")
	if email == "" {
		fmt.Print("To log into cyberspace, please enter your email:\n")
		//fmt.Scan(&email)
		fmt.Scan(&email)
	}
	fmt.Print("To sign in, please enter your password:\n")
	//fmt.Scan(&password)
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println("Error reading password:", err)
		os.Exit(1)
	}

	loginJson, err := json.Marshal(loginData{Email: email, Password: string(password)})
	if err != nil {
		fmt.Printf("Error encoding loginData to json: %s", err)
		os.Exit(1)
	}
	res, err := http.Post(url+"/auth/login", "application/json", bytes.NewBuffer(loginJson))
	//defer res.Body.Close()
	if err != nil {
		fmt.Printf("Error logging in: %s\n", err)
		os.Exit(1)
	}
	var authResp AuthResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&authResp)
	if err != nil {
		fmt.Printf("Error decoding json: %s\n", err)
		os.Exit(1)
	}
	return authResp.Data
}

type refreshData struct {
	RefreshToken string `json:"refreshToken"`
}

type refreshedTokens struct {
	Data struct {
		IDToken   string `json:"idToken"`
		RTDBToken string `json:"rtbdToken"`
	} `json:"data"`
}

func (c *APIClient) TokenRefresh() {

	refreshJson, err := json.Marshal(refreshData{RefreshToken: c.Tokens.RefreshToken})
	if err != nil {
		fmt.Printf("Error encoding refreshData to json: %s", err)
		os.Exit(1)
	}
	//fmt.Print(string(refreshJson))
	res, err := http.Post(c.ApiUrl+"/auth/refresh", "application/json", bytes.NewBuffer(refreshJson))
	if err != nil {
		fmt.Printf("Error refreshing auth tokens: %s\n", err)
		os.Exit(1)
	}
	//fmt.Println(res.Body, res.StatusCode)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status during token refresh: %d\n", res.StatusCode)
		os.Exit(1)
	}

	var refTokens refreshedTokens

	/*bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading body: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Raw response body: %s\n", string(bodyBytes))

	// Then decode from the bytes you already read
	err = json.Unmarshal(bodyBytes, &refTokens)*/

	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&refTokens)
	if err != nil {
		fmt.Printf("Error decoding json while refreshing tokens: %s\n", err)
		os.Exit(1)
	}
	//fmt.Print(refTokens)
	c.Tokens.IDToken = refTokens.Data.IDToken
	c.Tokens.RTDBToken = refTokens.Data.RTDBToken

	c.LastStatusCode = res.StatusCode
	if c.LastStatusCode == 401 {
		fmt.Print("Remedial Login.")
		c.Tokens = Login(c.ApiUrl)
	}

}

//////////////////////| Not yet implemented | Check Username availability & resend verification email |///////////////////////////

//

//

/////////////////////////| Unusable due to API access restrictions | Register |////////////////////////////////////

type registerData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Username string `json:"username"`
}

func Register(url string) AuthTokens { //client http.Client,
	var email string
	var password string
	var username string
	fmt.Print("To log into cyberspace, please enter your email:\n")
	fmt.Scan(&email)
	fmt.Print("To sign in, please enter your password:\n")
	fmt.Scan(&password)

	fmt.Print(`
	Please choose your username. The following rules apply:\n
	- 3-20 characters\n
	- Lowercase letters, numbers, underscores only\n
	- Cannot be a reserved name (admin, system, etc.)\n
	- Cannot contain prohibited words
	`)
	fmt.Scan(&username)

	loginJson, err := json.Marshal(registerData{Email: email, Password: password, Username: username})
	if err != nil {
		fmt.Printf("Error encoding registerData to json: %s", err)
		os.Exit(1)
	}
	res, err := http.Post(url+"/auth/register", "application/json", bytes.NewBuffer(loginJson))
	//defer res.Body.Close()
	if err != nil {
		fmt.Printf("Error logging in: %s\n", err)
		os.Exit(1)
	}
	var authResp AuthResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&authResp)
	if err != nil {
		fmt.Printf("Error decoding json: %s\n", err)
		os.Exit(1)
	}
	return authResp.Data
}
