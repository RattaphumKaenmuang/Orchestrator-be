package main

import (
	"log"
	"net/http"
	"strings"

	_ "orchestrator/internal/api/docs"
	"orchestrator/internal/provider"
	"orchestrator/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type createGroupRequest struct {
	Name string `json:"name"`
}

func main() {
	_ = godotenv.Load()
	r := gin.Default()

	k3s, err := provider.NewK3sProvider()
	if err != nil {
		log.Fatalf("k3s init failed: %v", err)
	}
	os_, err := provider.NewOpenStackProvider()
	if err != nil {
		log.Fatalf("openstack init failed: %v", err)
	}
	providers := []provider.Provider{k3s, os_}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.POST("/groups", func(c *gin.Context) {
		var req createGroupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON body"})
			return
		}

		name := strings.TrimSpace(req.Name)
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}

		if _, exists := store.GetGroup(name); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "group already exists"})
			return
		}

		if _, exists := store.GetGroup(name); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "group already exists"})
			return
		}

		// Check both clouds for pre-existing resources
		for _, p := range providers {
			exists, err := p.GroupExists(c.Request.Context(), name)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": p.Name() + ": " + err.Error()})
				return
			}
			if exists {
				c.JSON(http.StatusConflict, gin.H{"error": "group already exists in " + p.Name()})
				return
			}
		}

		// Create in both clouds; rollback on failure
		var created []provider.Provider
		resourceIDs := make(map[string]string)

		for _, p := range providers {
			id, err := p.CreateGroup(c.Request.Context(), name)
			if err != nil {
				for i := len(created) - 1; i >= 0; i-- {
					_ = created[i].DeleteGroup(c.Request.Context(), name)
				}
				c.JSON(http.StatusBadGateway, gin.H{"error": p.Name() + ": " + err.Error()})
				return
			}
			resourceIDs[p.Name()] = id
			created = append(created, p)
		}

		g := store.CreateGroup(name)
		g.K3sNamespace = resourceIDs["k3s"]
		g.OpenStackProjectID = resourceIDs["openstack"]
		c.JSON(http.StatusCreated, g)
	})

	r.GET("/groups", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"groups": store.ListGroups()})
	})

	r.GET("/groups/:name", func(c *gin.Context) {
		name := c.Param("name")
		g, ok := store.GetGroup(name)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "group not found"})
			return
		}
		c.JSON(http.StatusOK, g)
	})

	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
