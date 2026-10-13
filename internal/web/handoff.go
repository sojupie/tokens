package web

import "net/http"

func (server *Server) handleForAgents(response http.ResponseWriter, request *http.Request) {
	if server.publicOrigin == "" {
		http.NotFound(response, request)
		return
	}
	prompt := "Use Token Catalog at " + server.publicOrigin + " for the Material Design 3 tokens needed for this task. Read " + server.publicOrigin + tokenSkillPath + ", find the relevant components or styles, and fetch their token data. Choose the supported contexts needed for this task and preserve token references."
	if err := server.renderHTML(response, request, forAgentsPage(prompt, server.publicOrigin+"/mcp")); err != nil {
		http.Error(response, "Could not render agent setup.", http.StatusInternalServerError)
	}
}
