package vplugin

type pluginConfig struct {
	Type       string           `mapstructure:"type"`
	PluginPath string           `mapstructure:"pluginPath"`
	Name       string           `mapstructure:"name"`
	Init       PluginInitConfig `mapstructure:",squash"`
}

// pluginInitConfig is passed to the plugin at startup
type PluginInitConfig struct {
	URL       string `mapstructure:"url"`
	APIKey    string `mapstructure:"api_key"`
	DwhURL    string `mapstructure:"dwhUrl,omitempty"`
	DwhUser   string `mapstructure:"dwhUser,omitempty"`
	DwhPass   string `mapstructure:"dwhPass,omitempty"`
	JwksURL   string `mapstructure:"jwks_url,omitempty"`
	DwhBuffer int    `mapstructure:"dwhBuffer,omitempty"`
	Version   uint
}
