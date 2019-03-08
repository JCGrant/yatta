package notifiers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	gmail "google.golang.org/api/gmail/v1"
)

// GmailNotifier will send an email to the recipient with a subject as the notification message
type GmailNotifier struct {
	service   *gmail.Service
	recipient string
}

// NewGmailNotifier creates a GmailNotifier
func NewGmailNotifier(recipient string, credentialsPath string, tokenPath string) (*GmailNotifier, error) {
	b, err := ioutil.ReadFile(credentialsPath)
	if err != nil {
		return nil, errors.Wrapf(err, "reading credentials failed: %s", credentialsPath)
	}
	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return nil, errors.Wrap(err, "parsing credentials failed")
	}
	client, err := getClient(config, tokenPath)
	if err != nil {
		return nil, errors.Wrap(err, "getting client failed")
	}
	srv, err := gmail.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "creating Gmail client failed")
	}
	return &GmailNotifier{
		srv,
		recipient,
	}, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config, tokenPath string) (*http.Client, error) {
	tok, err := tokenFromFile(tokenPath)
	if err != nil {
		log.Printf("loading token failed: %s\n", tokenPath)
		log.Printf("fetching new token from web")
		tok, err = getTokenFromWeb(config, tokenPath)
		if err != nil {
			return nil, errors.Wrap(err, "fetching new token failed")
		}
		err := saveToken(tok, tokenPath)
		if err != nil {
			return nil, errors.Wrapf(err, "saving token failed: %s\n", tokenPath)
		}
	}
	return config.Client(context.Background(), tok), nil
}

// Retrieves a token from a local file.
func tokenFromFile(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config, tokenPath string) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, err
	}
	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, errors.Wrap(err, "retrieving token from web failed")
	}
	return tok, nil
}

// Saves a token to a file path.
func saveToken(token *oauth2.Token, path string) error {
	log.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.Wrapf(err, "opening path failed: %s", path)
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return errors.Wrap(err, "encoding token failed")
	}
	return nil
}

// Notify will send an email to n.recipient with message ad the subject
func (n *GmailNotifier) Notify(message string) error {
	return n.send(n.recipient, message, "")
}

func (n *GmailNotifier) send(recipient string, subject string, body string) error {
	message := fmt.Sprintf(`From:
To: %s
Subject: %s

%s`, recipient, subject, body)
	rawMessage := base64EncodeString(message)
	_, err := n.service.Users.Messages.Send("me", &gmail.Message{
		Raw: rawMessage,
	}).Do()
	if err != nil {
		return errors.Wrapf(err, "sending message failed: %s", message)
	}
	return nil
}

func base64EncodeString(message string) string {
	return string(base64.StdEncoding.EncodeToString([]byte(message)))
}
