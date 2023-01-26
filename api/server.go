package api

import (
	"github.com/gin-gonic/gin"
	"github.com/max-rodziyevsky/go-simple-bank/internal/repo"
)

type Server struct {
	store  *repo.Store
	router *gin.Engine
}

func NewServer(store *repo.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()
	err := router.SetTrustedProxies(nil)
	if err != nil {
		return nil
	}

	router.POST("/accounts", server.createAccount)
	router.GET("/accounts/:id", server.getAccount)
	router.GET("/accounts", server.listAccounts)
	router.PUT("/accounts", server.updateAccount)
	router.DELETE("accounts/:id", server.deleteAccount)

	server.router = router
	return server
}

func (s *Server) Start(address string) error {
	return s.router.Run(address)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
