package vplugin

// PluginInitConfig is passed to the plugin at startup
type PluginInitConfig = map[string]any
type pluginConfig struct {
	Init       PluginInitConfig `mapstructure:",remain"`
	Type       string           `mapstructure:"type"`
	PluginPath string           `mapstructure:"plugin_path"`
	Name       string           `mapstructure:"name"`
}
