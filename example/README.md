# When creating a new provider module

Start with creating a folder under the provider-folder with the provider name.

In [this folder](./example-game-provider/) you will find `example-game-provider` with some basic building blocks for creating a new provider module.
- `router.go` - set up provider wallet api as well as endpoints operator will call, such as gamelaunch
- `controller.go` - request parsing and validation from wallet calls
- `wallet_service.go` - mainly a mapping service converting the provider specific api to work with the Valkyrie PAM api
- `provider_service.go` - in the other way the Valkyrie operator api needs to be converted to work with the provider specifics, such as gamelaunch
- `config.go` - define auth and provider specific from the configuration.

Another step that needs to be done is defining what configuration is needed for this specific provider.
- `name` - Should be the same as `ProviderName` in router.go
- `base_path` - used to differentiate between providers within Valkyrie, used both when operator makes request toward valkyrie, as well as part of the url the provider should set up when making wallet requests toward the operator/Valkyrie
- `URL` - base path to game provider server for requests such as gamelaunch and render
- `auth` - "key - value" map for whatever auth fields are needed for the specific provider
- `provider_specific` - are there any other configuration that is needed for a specific provider? put them here.
## Looking at the code
[router.go](./example-game-provider/router.go) can be a good starting point. In there we set up the endpoints for the providers wallet api as well as the endpoints the operator will use to communicate with the game provider server.

It will register the routers with `provider.ProviderFactory()` and `provider.OperatorFactory()` in the `init`-function. In order for this to be executed, the game provider package needs to be imported in [routes/routes.go](../routes/routes.go).

``` go
import (
  // imports....
  _ "github.com/valkyrie-fnd/valkyrie/example/example-game-provider"
  _ "github.com/valkyrie-fnd/valkyrie/provider/your-new-game-provider-module"
  _ "github.com/valkyrie-fnd/valkyrie/provider/caleta"
  _ "github.com/valkyrie-fnd/valkyrie/provider/evolution"
  _ "github.com/valkyrie-fnd/valkyrie/provider/redtiger"
)
```

[wallet_service.go](./example-game-provider/wallet_service.go) is where you "convert" the game provider wallet api calls to fit within Valkyrie domain.

[provider_service.go](./example-game-provider/provider_service.go) is where you convert Valkyrie operator api to your own domain api.

