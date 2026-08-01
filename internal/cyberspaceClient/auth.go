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

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Unexpected status during token refresh: %s\n", res.Status)
		var errorResponse ErrorResponse
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&errorResponse)
		if err != nil {
			fmt.Print(err)
		}
		fmt.Print(errorResponse)
		os.Exit(1)
	}

	var authResp AuthResponse
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&authResp)
	if err != nil {
		fmt.Printf("Error decoding json: %s\n", err)
		os.Exit(1)
	}
	if (authResp.Data == AuthTokens{}) || authResp.Data.RefreshToken == "" {
		fmt.Println(authResp)
		fmt.Println("Error: Couldn't retrieve auth tokens. Exiting cyberspace...")
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

func (c *APIClient) TokenRefresh() error {
	if c.Tokens.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}
	refreshJson, err := json.Marshal(refreshData{RefreshToken: c.Tokens.RefreshToken})
	if err != nil {
		return fmt.Errorf("error encoding refresh data: %s", err)
	}
	res, err := http.Post(c.ApiUrl+"/auth/refresh", "application/json", bytes.NewBuffer(refreshJson))
	if err != nil {
		return fmt.Errorf("error refreshing auth tokens: %s", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		var errorResponse ErrorResponse
		if err := json.NewDecoder(res.Body).Decode(&errorResponse); err == nil && errorResponse.Error.Message != "" {
			return fmt.Errorf("token refresh failed: %s (%s)", errorResponse.Error.Message, errorResponse.Error.Code)
		}
		return fmt.Errorf("token refresh failed: %s", res.Status)
	}

	var refTokens refreshedTokens
	if err := json.NewDecoder(res.Body).Decode(&refTokens); err != nil {
		return fmt.Errorf("error decoding token refresh response: %s", err)
	}
	c.Tokens.IDToken = refTokens.Data.IDToken
	c.Tokens.RTDBToken = refTokens.Data.RTDBToken
	c.LastStatusCode = res.StatusCode
	return nil
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
