package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/oxtyped/gpodder2go/pkg/apis"
	"github.com/oxtyped/gpodder2go/pkg/data"
	"github.com/oxtyped/gpodder2go/pkg/store"

	m2 "github.com/oxtyped/gpodder2go/pkg/middleware"
	"github.com/spf13/cobra"
)

var database string
var addr string

func init() {
	serveCmd.Flags().StringVarP(&database, "database", "d", "g2g.db", "filename of sqlite3 database to use")
	serveCmd.Flags().StringVarP(&addr, "addr", "b", "localhost:3005", "ip:port for server to be binded to")
	rootCmd.AddCommand(serveCmd)
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start gpodder2go server",
	Run: func(cmd *cobra.Command, args []string) {

		verifierSecretKey := os.Getenv("VERIFIER_SECRET_KEY")

		if verifierSecretKey == "" {
			fmt.Println("VERIFIER_SECRET_KEY is missing")
			return
		}

		r := chi.NewRouter()
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)

		store := store.NewCacheStore()

		// take in db flag and parse it
		dataInterface := data.NewSQLite(database)
		deviceAPI := apis.DeviceAPI{Store: store, Data: dataInterface}
		subscriptionAPI := apis.SubscriptionAPI{Data: dataInterface}
		episodeAPI := apis.EpisodeAPI{Data: dataInterface}
		userAPI := apis.NewUserAPI(dataInterface, verifierSecretKey)

		// TODO: Add the authentication middlewares for the various places

		// auth
		r.Group(func(r chi.Router) {
			r.Post("/api/2/auth/{username}/login.json", userAPI.HandleLogin)
		})

		r.Group(func(r chi.Router) {
			r.Use(m2.Verifier(verifierSecretKey))
			r.Post("/api/internal/users", userAPI.HandleUserCreate)

			// device
			r.Post("/api/2/devices/{username}/{deviceid}.json", deviceAPI.HandleUpdateDevice)
			r.Get("/api/2/devices/{username}.json", deviceAPI.HandleGetDevices)

			// subscriptions
			r.Get("/api/2/subscriptions/{username}/{deviceid}.{format}", subscriptionAPI.HandleGetDeviceSubscriptionChange)
			//	r.Put("/api/2/subscriptions/{username}/{deviceid}.{format}", subscriptionAPI.HandleUploadDeviceSubscription)
			r.Post("/api/2/subscriptions/{username}/{deviceid}.{format}", subscriptionAPI.HandleUploadDeviceSubscriptionChange)

			r.Put("/subscriptions/{username}/{deviceid}.{format}", subscriptionAPI.HandleUploadDeviceSubscription)
			r.Get("/subscriptions/{username}/{deviceid}.{format}", subscriptionAPI.HandleGetDeviceSubscription)
			r.Get("/subscriptions/{username}.{format}", subscriptionAPI.HandleGetSubscription)

			// episodes
			r.Get("/api/2/episodes/{username}.{format}", episodeAPI.HandleEpisodeAction)
			r.Post("/api/2/episodes/{username}.{format}", episodeAPI.HandleUploadEpisodeAction)

			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("welcome"))
			})
		})

		log.Printf("💻 Starting server at %s", addr)
		err := http.ListenAndServe(addr, r)
		if err != nil {
			log.Fatal(err)
		}

	},
}
