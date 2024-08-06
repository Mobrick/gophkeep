package handler

import (
	"gophkeep/internal/config"
	"gophkeep/internal/database"
)

type Env struct {
	ConfigStruct *config.Config
	Storage      database.Storage
}
