package service

import (
	"log"
	"net/http"
)

func (h *Service) GetVersion(r *http.Request, args *struct{ Who string }, reply *struct{ Message string }) error {
	log.Println("GetVersion", args.Who)
	reply.Message = "Hello, " + args.Who + "!"
	return nil
}
