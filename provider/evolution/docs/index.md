
### Games

[Evolution](https://www.evolution.com/) offers leading live games such as casino, cards & craps.

### Evolution API

See evolution API (https://studio.evolution.com/api/docs) for details on Evolution eco system.

### Valkyrie integration

Integrate your gaming lobby and wallet system (often enough referred to as "PAM") to Valkyrie and you will be able to access all games offered by Evolution Gaming!

You can either integrate directly to the [Valkyrie standardized gaming API](/docs/wallet/valkyrie-pam-api)  or implement a proprietary integration to Valkyrie. If you choose proprietary option, you can either add the solution to the valkyrie repo, or make use of the vplugin option.

### Configuration

```yaml
providers:
  - name: Evolution
    url: 'https://Evo-baseurl' # 
    auth:
      casino_key: ${EVO_CASINO_KEY} # Casino specific key provided by Evolution
      api_key: ${EVO_API_KEY} # Evolution api key
      casino_token:  ${EVO_CASINO_API_TOKEN} # Token used for game launch requests toward evolution backend
```