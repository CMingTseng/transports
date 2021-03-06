package main

import (
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/matiasinsaurralde/transports"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	godotenv.Load()

	fmt.Println("Transports test (Whatsapp/Yowsup)")

	whatsappTransport := transports.WhatsappTransport{
		Login:             os.Getenv("WA_SERVER_LOGIN"),
		Password:          os.Getenv("WA_SERVER_PASSWORD"),
		Contact:           os.Getenv("WA_SERVER_CONTACT"),
		UseTor: true,
		YowsupWrapperPort: "8889",
	}

	whatsappTransport.Listen(func(t *transports.WhatsappTransport) {
		for _, Value := range t.Messages {
			request := t.Serializer.DeserializeRequest([]byte(Value.Body))
			if request.Method == "" {
				fmt.Println("*** Ignoring message", "\n")
				t.PurgeMessage(Value.Id)
				return
			}

			fmt.Println("--> Receiving, accepting message\n", request, "\n")

			transport := &http.Transport{}
			if t.UseTor {
				transport.Dial = transports.TorDialer().Dial
			}
			client := &http.Client{ Transport: transport}
			response, _ := client.Do(request)
			defer response.Body.Close()

			rawBody, _ := ioutil.ReadAll(response.Body)

			serializedResponse := t.Serializer.Serialize(response, false).(transports.Response)
			serializedResponse.Body = string(rawBody)

			jsonResponse, _ := json.Marshal(serializedResponse)

			t.SendMessage(string(jsonResponse))

			t.PurgeMessage(Value.Id)

		}

		t.Messages = make([]transports.WhatsappMessage, 0)
	})
}
