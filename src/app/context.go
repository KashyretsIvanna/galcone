package app

import (
	"galcone/src/galcone/matchmaking"
	"galcone/src/galcone/wsctx"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

// App has router and db instances
type GlobalContext struct {
	Router             *mux.Router
	UserRepository     matchmaking.UserRepository
	GameRoomRepository matchmaking.GameRoomRepository
	// Temporary
	Hub *wsctx.Hub
}

type METHOD string

const (
	GET    METHOD = "GET"
	POST   METHOD = "POST"
	PUT    METHOD = "PUT"
	DELETE METHOD = "DELETE"
)

type RestEndpoint struct {
	URL     string
	Method  METHOD
	Handler func(*GlobalContext, http.ResponseWriter, *http.Request)
}

type WSEndpoint struct {
	URL     string
	Handler func(*GlobalContext, http.ResponseWriter, *http.Request)
}

func (ep *WSEndpoint) AsHandler(ctx *GlobalContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ep.Handler(ctx, w, r)
	})
}

// Initialize initializes the app with predefined configuration
func (ctx *GlobalContext) Initialize() {
	ctx.Hub = wsctx.NewHub()
	go ctx.Hub.Run()

	ctx.UserRepository = matchmaking.UserRepoDummyImpl()
	ctx.GameRoomRepository = matchmaking.GameRoomDummyImpl()

	ctx.Router = mux.NewRouter()
}

// Initialize initializes the app with predefined configuration
func (ctx *GlobalContext) InitializeDummy() {
	ctx.Hub = wsctx.NewHub()
	go ctx.Hub.Run()

	ctx.UserRepository = matchmaking.UserRepoDummyImpl()
	ctx.GameRoomRepository = matchmaking.GameRoomDummyImpl()

	ctx.Router = mux.NewRouter()
}

func (ctx *GlobalContext) SetRestAPI(routes *[]*RestEndpoint) {
	for _, r := range *routes {
		ctx.Router.HandleFunc(r.URL, func(wr http.ResponseWriter, req *http.Request) {
			r.Handler(ctx, wr, req)
		}).Methods(string(r.Method))
	}
}

func (ctx *GlobalContext) SetSocketAPI(routes *[]*WSEndpoint) {
	for _, r := range *routes {
		ctx.Router.Handle(r.URL, r.AsHandler(ctx))
	}
}

func (ctx *GlobalContext) Run() {
	log.SetFlags(0)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), ctx.Router))
}
