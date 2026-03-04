package main

import "orchestrator/internal/model"

// @title Orchestrator API
// @version 1.0
// @description Demo API for multi-cloud Group and Instance orchestration.
// @BasePath /

// ErrorResponse is a generic error payload for docs.
type ErrorResponse struct {
	Error string `json:"error"`
}

// GroupListResponse matches GET /groups response shape.
type GroupListResponse struct {
	Groups []*model.Group `json:"groups"`
}

// createGroupDocs godoc
// @Summary Create a group
// @Description Creates a new Group (namespace in k3s + project in OpenStack). Rolls back on partial failure.
// @Tags groups
// @Accept json
// @Produce json
// @Param payload body createGroupRequest true "Group payload"
// @Success 201 {object} model.Group
// @Failure 400 {object} ErrorResponse "Invalid or missing name"
// @Failure 409 {object} ErrorResponse "Group already exists"
// @Failure 502 {object} ErrorResponse "Cloud provider error"
// @Router /groups [post]
func createGroupDocs() {}

// listGroupsDocs godoc
// @Summary List groups
// @Description Returns all Groups.
// @Tags groups
// @Produce json
// @Success 200 {object} GroupListResponse
// @Router /groups [get]
func listGroupsDocs() {}

// getGroupDocs godoc
// @Summary Get group by name
// @Description Returns one Group by its name, including its instances.
// @Tags groups
// @Produce json
// @Param name path string true "Group name"
// @Success 200 {object} model.Group
// @Failure 404 {object} ErrorResponse "Group not found"
// @Router /groups/{name} [get]
func getGroupDocs() {}

// createInstanceDocs godoc
// @Summary Create an instance in a group
// @Description Creates a new Instance. GPU > 0 routes to k3s (pod), otherwise to OpenStack (server). Flavor is auto-matched.
// @Tags instances
// @Accept json
// @Produce json
// @Param name path string true "Group name"
// @Param payload body createInstanceRequest true "Instance specs"
// @Success 201 {object} model.Instance
// @Failure 400 {object} ErrorResponse "Invalid or missing fields"
// @Failure 404 {object} ErrorResponse "Group not found"
// @Failure 500 {object} ErrorResponse "Unknown provider"
// @Failure 502 {object} ErrorResponse "Cloud provider error"
// @Router /groups/{name}/instances [post]
func createInstanceDocs() {}
