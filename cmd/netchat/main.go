package main

import (
	"github.com/cfanatic/go-netchat/internal/database"
	"github.com/cfanatic/go-netchat/internal/restapi"
	"github.com/cfanatic/go-netchat/internal/settings"
)

func main() {
	settings.Get(settings.PORT)
	restapi.New()
	database.New()
}
