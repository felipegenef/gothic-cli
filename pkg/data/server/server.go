{{.MainServerPackageName}}

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"{{.GoModName}}/src/routes"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

func {{.MainServerFunctionName}} {
	godotenv.Load()
	var localServe = os.Getenv("LOCAL_SERVE")
	var isLocal = len(localServe) > 0 && localServe == "true"
	router := chi.NewMux()
	router.Use(middleware.Logger)

	/**
	*                              Public assets folder
	*
	* Here is where you serve your static files for your application like css files,
	* javascript files, images, videos etc. If in AWS there is an origin configured for
	* AWS Cloudfront that will serve those files from an s3 bucket and cache them in the edge.
	* If you are running this program locally with the "make hot-reload" command, the files
	* will be served from your local public folder. To control local file serving behaviour
	* change the LOCAL_SERVE environment variable to "false" on your .env file
	*
	*
	 */
	if isLocal {
		slog.Info("application serving local public folder")
		router.Handle("/public/*", http.StripPrefix("/public/", http.FileServer(http.Dir("./public/"))))
	}

	router.Group(routes.RegisterFileBasedRoutes)

	port := os.Getenv("HTTP_LISTEN_ADDR")
	slog.Info("application running", "port", port)
	log.Fatal(http.ListenAndServe(port, router))
}
