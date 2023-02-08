# When creating a new provider module

Start with creating a folder under the provider-folder with the provider name.

In [this folder](./example-game-provider/) you will find `example-game-provider` with some basic building blocks for creating a new provider module.
- `router.go` - set up provider wallet api as well as endpoints operator will call, such as gamelaunch
- `controller.go` - request parsing and validation from wallet calls
- `wallet_service.go` - mainly a mapping service converting the provider specific api to work with the Valkyrie PAM api
- `provider_service.go` - in the other way the Valkyrie operator api needs to be converted to work with the provider specifics, such as gamelaunch
- `config.go` - define "auth" and "provider_specific" from the configuration.
- `models.go` - define any models used by the provider wallet api, such as request models.

Another step that needs to be done is defining what configuration is needed for this specific provider.
- `name` - Should be the same as `ProviderName` in router.go
- `base_path` - used to differentiate between providers within Valkyrie, used both when operator makes request toward valkyrie, as well as part of the url the provider should set up when making wallet requests toward the operator/Valkyrie
- `URL` - base path to game provider server for requests such as gamelaunch and render
- `auth` - "key - value" map for whatever auth fields are needed for the specific provider
- `provider_specific` - are there any other configuration that is needed for a specific provider? put them here.

See [caleta configuration](../provider/caleta/docs/index.md) for reference, as well as the [website](https://valbyrie.bet/docs/get-started/configuration) documentation.

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

## Documentation

Adding a subfolder to the provider called "docs", will enable the valkyrie site to pick up the provider module and add the information to it. Any `.md` or `.mdx` files will be added to the the documentation of the provider on [valkyrie.bet](https://valkyrie.bet).

Any images or other assets should be placed in the assets sub-folder.

Add a `config.yml` with info that will be used in the [providers](https://valbyrie.bet/providers) section of the website.

```yaml
id: 99 # just increase from other existing ids
name: "Example Games" # Name of provider shown on the website
cardImage: "example-logo.png" # Image shown in the provider card. The images should be placed in the assets folder
description: "Amazing games from example provider" # Short description that will be shown in the provider card
path: "exampleprovider" # url path to provider on the valkyrie website
```

See Caleta, Evolution and Red tiger for reference.