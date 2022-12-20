### Games

[Red Tiger](https://www.redtiger.com/) offers leading games in the slot space.

### Red Tiger API

See Red Tigers API (https://dev.redtigergaming.com/#!/) for details on Red Tiger eco system.

### Valkyrie integration

Integrate your gaming lobby and wallet system (often enough referred to as "PAM") to Valkyrie and you will be able to access all games offered by Red Tiger!

You can either integrate directly to the [Valkyrie standardized gaming API](/docs/wallet/valkyrie-pam-api)  or implement a proprietary integration to Valkyrie. If you choose proprietary option, you can either add the solution to the valkyrie repo, or make use of the vplugin option.

### Configuration

```yaml
providers:
  - name: Red Tiger
    url: 'https://redtiger' # see your settings page at https://dev.redtigergaming.com to see your operator specific url
    auth:
      api_key: ${RED_TIGER_API_KEY} # Red tiger api key. Found in https://dev.redtigergaming.com
      recon_token:  ${RECON_TOKEN} # Reconciliation token. See https://dev.redtigergaming.com
```