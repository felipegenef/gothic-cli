package server

import (
	"embed"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/felipegenef/gothic-cli/src/api"
	"github.com/felipegenef/gothic-cli/src/components"
	"github.com/felipegenef/gothic-cli/src/pages"
	handler "github.com/felipegenef/gothic-cli/src/utils"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
)

//go:embed public
var FS embed.FS
var isLocal bool

func startServer() {
	godotenv.Load()

	router := chi.NewMux()
	localServe := os.Getenv("LOCAL_SERVE")
	isLocal = len(localServe) > 0 && localServe == "true"
	revalidateLocalTime := time.Now()

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
		router.Handle("/*", http.FileServer(http.FS(FS)))
	}

	/**
	*                         Page routes
	*
	* Here is where you serve your page routes.You can add how many as you want.
	* Just render and serve them as html with templ templating as shown below.
	* For more info check templ documentation:
	*
	*                         https://templ.guide/
	*
	 */
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handler.Render(r, w, pages.Index())
	})
	/**
	*                              ISR Page Example
	*
	* Here is one example of how you can have the same behaviour as
	* Next.js Incremental Static Regeneration (ISR) pages. You can
	* Do the same with json routes as well as component routes.
	* Just need to set the header Cache-Control with your desired behaviour
	*
	*
	 */
	router.Get("/cachedPageRoute", func(w http.ResponseWriter, r *http.Request) {
		currentTime := time.Now()
		/**
		* Local ISR can be removed if desired. Adding this piece of code
		* makes caching work on local serving but that is not needed for
		* production. If you wish just keep the else statement.
		 */
		if isLocal {
			// Local code example of revalidate caching
			if currentTime.Sub(revalidateLocalTime) > 10*time.Second {
				revalidateLocalTime = currentTime
			}
			handler.Render(r, w, pages.Revalidate(revalidateLocalTime))
		} else {
			// Max cache age for CloudFront is 31536000 = 1 year
			w.Header().Set("Cache-Control", "max-age=10, stale-while-revalidate=10, stale-if-error=10")
			handler.Render(r, w, pages.Revalidate(currentTime))
		}

	})
	/**
	*                            Component ROUTES
	*
	* For HTMX to work, you need to add routes with the desired behaviour
	* of your application. For more info check HTMX docummentation:
	*
	*                            https://htmx.org/
	*
	*
	 */
	router.Get("/optimizedImage/{name}/{extension}", func(w http.ResponseWriter, r *http.Request) {
		imgName := chi.URLParam(r, "name")
		imgExtension := chi.URLParam(r, "extension")
		handler.Render(r, w, components.OptimizedImage(false, imgName, imgExtension))
	})
	/**
	*                                 API ROUTES
	*
	* Those are the routes you return json files for external clients outside your app.
	*
	*
	 */
	router.Get("/helloWorld", api.HelloWorld)

	port := os.Getenv("HTTP_LISTEN_ADDR")
	slog.Info("application running", "port", port)
	log.Fatal(http.ListenAndServe(port, router))
}
