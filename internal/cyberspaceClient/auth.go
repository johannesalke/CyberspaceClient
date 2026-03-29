package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	var password string
	fmt.Print("To log into cyberspace, please enter your email:\n")
	//fmt.Scan(&email)
	email = "not0really0anonymous@gmail.com"
	fmt.Print("To sign in, please enter your password:\n")
	fmt.Scan(&password)

	loginJson, err := json.Marshal(loginData{Email: email, Password: password})
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
	IDToken   string `json:"idToken"`
	RTDBToken string `json:"rtbdToken"`
}

func TokenRefresh(url string, tokens AuthTokens) AuthTokens {
	refreshJson, err := json.Marshal(refreshData{RefreshToken: tokens.RefreshToken})
	if err != nil {
		fmt.Printf("Error encoding refreshData to json: %s", err)
		os.Exit(1)
	}
	res, err := http.Post(url+"/auth/refresh", "application/json", bytes.NewBuffer(refreshJson))
	//defer res.Body.Close()
	if err != nil {
		fmt.Printf("Error refreshing auth tokens: %s\n", err)
		os.Exit(1)
	}
	var refTokens refreshedTokens
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&refTokens)
	if err != nil {
		fmt.Printf("Error decoding json: %s\n", err)
		os.Exit(1)
	}
	tokens.IDToken = refTokens.IDToken
	tokens.RTDBToken = refTokens.RTDBToken
	return tokens
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
