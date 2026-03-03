package main

import "orchestrator/internal/model"

// @title Orchestrator API
// @version 1.0
// @description Demo API for Group creation and inspection.
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
// @Description Creates a new Group.
// @Tags groups
// @Accept json
// @Produce json
// @Param payload body createGroupRequest true "Group payload"
// @Success 201 {object} model.Group
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
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
// @Description Returns one Group by its name.
// @Tags groups
// @Produce json
// @Param name path string true "Group name"
// @Success 200 {object} model.Group
// @Failure 404 {object} ErrorResponse
// @Router /groups/{name} [get]
func getGroupDocs() {}
