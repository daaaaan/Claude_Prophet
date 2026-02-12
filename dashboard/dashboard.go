package dashboard

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

//go:embed static/templates/*.html static/css/*.css
var embeddedFiles embed.FS

// Register sets up dashboard routes on the provided gin router.
// It configures the HTML template rendering from embedded files,
// serves static assets, and registers the WebSocket endpoint.
func Register(router *gin.Engine, hub *Hub) {
	// Parse templates from embedded filesystem
	templ := template.Must(template.New("").ParseFS(embeddedFiles, "static/templates/*.html"))
	router.SetHTMLTemplate(templ)

	// Create sub-filesystem for static assets
	staticFS, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatalf("Failed to create static sub-filesystem: %v", err)
	}

	// Serve static assets
	router.StaticFS("/dashboard/static", http.FS(staticFS))

	// Dashboard page handler
	router.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{})
	})

	// WebSocket handler
	router.GET("/ws", func(c *gin.Context) {
		serveWs(hub, c.Writer, c.Request)
	})

	log.Info("Dashboard routes registered: /dashboard, /ws, /dashboard/static/*")
}
