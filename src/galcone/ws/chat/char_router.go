package chat

import "galcone/src/app"

var Routes = []*app.WSEndpoint{
    {
        URL:     "/",
        Handler: ServeHome,
    },
    {
        URL: "/ws",
        Handler: ServeWs,
    },
}

